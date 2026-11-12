import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { restoreConfig, sql } from '../../lib/db';
import { field, fillTranslated, go, open, openSettings, panel, psWindow, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { users: [] as string[], prizes: [] as string[], types: [] as string[], rules: [] as string[], metaRules: [] as string[], tables: [] as string[], streaks: [] as string[], metas: [] as string[] };
let prize: any;
let P: Player;

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
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

async function selectIn(scope: Locator, label: string, value: string) {
  await field(scope, label).locator('select').selectOption(value);
}

async function playerPage(page: Page, id: string) {
  await page.goto(CONSOLE_URL + '/#/players/' + id);
  await expect(page.locator('[data-player-progress] h2')).toBeVisible({ timeout: 20_000 });
  await expect(page.locator('[data-player-progress] .animate-pulse')).toHaveCount(0);
}

test.beforeAll(async () => {
  prize = await eco.post('/prizes', { key: 'e2e_' + k + '_coins', name: { fa: 'جایزهٔ آزمایشی ' + k, en: 'Test prize ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 7 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  P = await devicePlayer('e2e-ev-' + k, ('ev' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_reward_claim WHERE prize_id='${prize.id}'; DELETE FROM economy_user_prize WHERE prize_id='${prize.id}'`);
  for (const id of made.rules) await eco.del('/event-rules/' + id);
  for (const id of made.metaRules) await eco.del('/meta-rules/' + id);
  for (const id of made.tables) await eco.del('/level-tables/' + id);
  for (const id of made.streaks) await eco.del('/streaks/' + id);
  for (const id of made.types) await eco.del('/event-types/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
});

test.describe('events: definitions', () => {
  test('an event type made in the console is accepted from the game within its range; built-in events are locked to the server', async ({ page }) => {
    await open(page, '#/economy/event-types/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('Bad Key');
    await fillTranslated(p, 'name', { fa: 'کلیک ' + k });
    await savePanel(page).catch(() => null);
    await expect(field(p, 'کلید')).toContainText('حرف کوچک لاتین');
    const key = 'e2e_' + k + '_tap';
    await field(p, 'کلید').locator('input').fill(key);
    await fillTranslated(p, 'name', { fa: 'کلیک ' + k, en: 'Tap ' + k });
    await p.getByText('بازی هم بتواند ثبتش کند').click();
    await field(p, 'کمترین مقدار').locator('input').fill('۱');
    await field(p, 'بیشترین مقدار').locator('input').fill('5');
    await field(p, 'سقف برای هر بازیکن در دقیقه').locator('input').fill('۳');
    await savePanel(page);
    const type = await byKey('event-types', key);
    made.types.push(type.id);
    expect(type).toMatchObject({ clientAllowed: true, valueMin: 1, valueMax: 5, rateLimit: 3, isActive: true, isSystem: false });
    await search(page, key);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('بازی هم');
    await expect(rows(page).first()).toContainText('۳ در دقیقه');

    expect((await rpc(P, 'economy_event_record', { key, value: 9 })).data.message).toBe('VALUE_OUT_OF_RANGE');
    expect((await rpc(P, 'economy_event_record', { key: 'login' })).data.message).toBe('EVENT_NOT_ALLOWED');
    expect((await rpc(P, 'economy_event_record', { key: 'e2e_nope_' + k })).data.message).toBe('EVENT_TYPE_NOT_FOUND');
    for (let i = 0; i < 3; i++) expect((await rpc(P, 'economy_event_record', { key, value: 2 })).status).toBe(200);
    expect((await rpc(P, 'economy_event_record', { key, value: 2 })).data.message).toBe('TOO_MANY_EVENTS');

    await go(page, '#/economy/event-types?q=login');
    const login = rows(page).filter({ hasText: 'login' }).first();
    await expect(login).toContainText('داخلی');
    await expect(login).toContainText('فقط سرور');
    await login.click();
    await settlePanel(page);
    await expect(panel(page)).toContainText('این رویداد داخلی است');
    await expect(panel(page).getByText('بازی هم بتواند ثبتش کند')).toBeHidden();
    await expect(panel(page).locator('button', { hasText: 'حذف' })).toBeVisible();
    await panel(page).locator('button', { hasText: 'حذف' }).click();
    await panel(page).locator('button', { hasText: 'بله، حذف شود' }).click();
    await expect(panel(page)).toContainText('رویدادهای داخلی حذف نمی‌شوند');
  });

  test('an event prize rule shows a live sentence, refuses impossible repeats and pays automatically', async ({ page }) => {
    const key = 'e2e_' + k + '_tap';
    await open(page, '#/economy/event-rules/new');
    await settlePanel(page);
    const p = panel(page);
    const ruleKey = 'e2e_' + k + '_rule';
    await field(p, 'کلید').locator('input').fill(ruleKey);
    await fillTranslated(p, 'name', { fa: 'دو کلیک در روز', en: 'Two taps a day' });
    await pickIn(field(p, 'رویداد'), key);
    await selectIn(p, 'شمارش', 'max');
    await selectIn(p, 'تکرار', 'every');
    await field(p, 'هدف').locator('input').fill('۲');
    await pickIn(field(p, /^جایزه$/), 'e2e_' + k + '_coins');
    await savePanel(page).catch(() => null);
    await expect(field(p, 'تکرار')).toContainText('با تعداد یا جمع کار می‌کند');
    await selectIn(p, 'شمارش', 'count');
    await selectIn(p, 'تکرار', 'period');
    await selectIn(p, 'روش دادن', 'auto');
    await expect(p.locator('[data-rule-preview]')).toContainText('«کلیک ' + k + '» در یک روز ۲ بار · هر روز یک بار');
    await savePanel(page);
    const rule = await byKey('event-rules', ruleKey);
    made.rules.push(rule.id);
    expect(rule).toMatchObject({ measure: 'count', period: 'day', repeat: 'period', threshold: 2, grantMode: 'auto' });
    await search(page, ruleKey);
    await expect(rows(page).first().locator('[data-rule-sentence]')).toContainText('در یک روز ۲ بار');
    await expect(rows(page).first()).toContainText('خودکار');

    const other = await devicePlayer('e2e-ev2-' + k);
    made.users.push(other.id);
    const before = (await rpc(other, 'economy_balances', {})).data.wallet.coin || 0;
    await rpc(other, 'economy_event_record', { key, value: 1 });
    const res = await rpc(other, 'economy_event_record', { key, value: 1 });
    expect(res.data.claims).toHaveLength(1);
    expect(res.data.claims[0]).toMatchObject({ status: 'granted', grantMode: 'auto' });
    expect((await rpc(other, 'economy_balances', {})).data.wallet.coin).toBe(before + 7);
    const progress = (await rpc(other, 'economy_events_progress', { lang: 'fa' })).data.items.find((x: any) => x.key === ruleKey);
    expect(progress).toMatchObject({ current: 2, done: true, claims: 1, name: 'دو کلیک در روز' });
    await page.reload();
    await root(page).waitFor();
    await expect(rows(page).first().locator('td').nth(3)).toHaveText('۱');
    const del = await eco.del('/event-rules/' + rule.id);
    expect(del.data.error).toBe('IN_USE');
  });

  test('meta rule, level table rows and a streak with rewards, built in their editors', async ({ page }) => {
    const xp = await eco.post('/metas', { key: 'e2e_' + k + '_xp', name: { fa: 'تجربهٔ ' + k }, defaultValue: 0, serverOnly: false, settings: {} });
    const lvl = await eco.post('/metas', { key: 'e2e_' + k + '_lvl', name: { fa: 'سطح ' + k }, defaultValue: 1, serverOnly: false, settings: {} });
    made.metas.push(xp.id, lvl.id);
    const key = 'e2e_' + k + '_tap';

    await open(page, '#/economy/meta-rules/new');
    await settlePanel(page);
    let p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_mr');
    await fillTranslated(p, 'name', { fa: 'تجربه از کلیک' });
    await pickIn(field(p, 'رویداد'), key);
    await pickIn(field(p, 'متا'), 'e2e_' + k + '_xp');
    await field(p, 'عدد').locator('input').fill('۵۰');
    await savePanel(page);
    const mr = await byKey('meta-rules', 'e2e_' + k + '_mr');
    made.metaRules.push(mr.id);
    expect(mr).toMatchObject({ action: 'add', valueSource: 'event', amount: 50 });

    await go(page, '#/economy/level-tables/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_lv');
    await fillTranslated(p, 'name', { fa: 'سطح‌های آزمایشی' });
    await pickIn(field(p, 'متای امتیاز'), 'e2e_' + k + '_xp');
    await pickIn(field(p, 'سطح در این متا نوشته شود'), 'e2e_' + k + '_lvl');
    const steps = p.locator('[data-rows="steps"]');
    await expect(steps.locator('[data-row]')).toHaveCount(2);
    await steps.locator('[data-row="1"] [data-col="threshold"] input').fill('۱۰۰');
    await steps.locator('[data-rows-add="steps"]').click();
    await steps.locator('[data-row="2"] [data-col="threshold"] input').fill('50');
    await savePanel(page).catch(() => null);
    await expect(steps.locator('[data-row="2"] .eco-items-err')).toContainText('آستانهٔ برابر یا بیشتر');
    await steps.locator('[data-row="2"] [data-col="threshold"] input').fill('300');
    await pickIn(steps.locator('[data-row="2"] [data-col="prizeId"]'), 'e2e_' + k + '_coins');
    await steps.locator('[data-row="2"] [data-col="grantMode"] select').selectOption('auto');
    await steps.locator('[data-row="1"] [data-col="name"] input').first().fill('قهرمان');
    await savePanel(page);
    const table = await byKey('level-tables', 'e2e_' + k + '_lv');
    made.tables.push(table.id);
    expect(table.steps.map((s: any) => [s.level, s.threshold, s.grantMode])).toEqual([[1, 0, 'claim'], [2, 100, 'claim'], [3, 300, 'auto']]);
    expect(table.steps[1].name.fa).toBe('قهرمان');
    await search(page, 'e2e_' + k + '_lv');
    await expect(rows(page).first()).toContainText('۳ سطح');

    await go(page, '#/economy/streaks/new');
    await settlePanel(page);
    p = panel(page);
    await expect(p).toContainText('به Hiro نیازی ندارد');
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_sk');
    await fillTranslated(p, 'name', { fa: 'سه روز کلیک' });
    await pickIn(field(p, 'رویداد شمارنده'), key);
    await field(p, 'طول دوره').locator('input').fill('۳');
    await field(p, 'روزهای فرصت جبران').locator('input').fill('1');
    const rewards = p.locator('[data-rows="rewards"]');
    await rewards.locator('[data-rows-add="rewards"]').click();
    await pickIn(rewards.locator('[data-row="0"] [data-col="prizeId"]'), 'e2e_' + k + '_coins');
    await rewards.locator('[data-rows-add="rewards"]').click();
    await rewards.locator('[data-rows-add="rewards"]').click();
    await rewards.locator('[data-row="2"] [data-col="day"] input').fill('5');
    await savePanel(page).catch(() => null);
    await expect(rewards.locator('[data-row="2"] .eco-items-err')).toContainText('بعد از پایان دوره');
    await rewards.locator('[data-row="2"] [data-col="day"] input').fill('3');
    await savePanel(page);
    const streak = await byKey('streaks', 'e2e_' + k + '_sk');
    made.streaks.push(streak.id);
    expect(streak).toMatchObject({ periodDays: 1, gracePeriods: 1, cycleLength: 3, missAction: 'reset' });
    expect(streak.rewards.map((r: any) => r.day)).toEqual([1, 2, 3]);

    const r = await rpc(P, 'economy_event_record', { key, value: 3 });
    expect(r.status === 200 || r.data.message === 'TOO_MANY_EVENTS').toBe(true);
  });
});

test.describe('events: settings, log and the player card', () => {
  test('event settings validate and show the clock of the chosen zone', async ({ page }) => {
    try {
      await openSettings(page, 'events');
      const body = psWindow(page);
      await expect(body.locator('[data-zone-now]')).toContainText('اکنون در این منطقه');
      await field(body, 'منطقهٔ زمانی').locator('select').selectOption('UTC');
      await field(body, 'شروع هفته').locator('select').selectOption('1');
      await field(body, 'نگه‌داری گزارش (روز)').locator('input').fill('0');
      await body.locator('[data-settings-save]').click();
      await expect(field(body, 'نگه‌داری گزارش (روز)')).toContainText('عدد');
      await field(body, 'نگه‌داری گزارش (روز)').locator('input').fill('۴۵');
      await body.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toBeVisible();
      const s = await eco.get('/event-settings');
      expect(s).toMatchObject({ timezone: 'UTC', weekStart: 1, retentionDays: 45 });
      const bad = await consoleApi('PUT', '/v2/console/economy/event-settings', { timezone: 'Mars/Base', weekStart: 1, retentionDays: 45, playerRateLimit: 100 });
      expect(bad.data.error).toBe('INVALID_TIMEZONE');
    } finally {
      restoreConfig();
    }
  });

  test('the player page shows progress; record an event, reset the streak and follow the log', async ({ page }) => {
    const key = 'e2e_' + k + '_tap';
    await playerPage(page, P.id);
    const card = page.locator('[data-player-progress]');
    await expect(card).toContainText('پیشرفت بازیکن');
    await expect(card.locator(`[data-pp-total="${key}"]`)).toBeVisible();
    await card.locator('[data-record-event]').click();
    const dlg = page.locator('[data-form-dialog="record-event"]');
    await expect(dlg).toBeVisible();
    await pickIn(dlg, key);
    await dlg.locator('input[inputmode="decimal"]').fill('۲');
    await dlg.locator('[data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('رویداد ثبت شد');
    const lv = card.locator(`[data-pp-level="e2e_${k}_lv"]`);
    await expect(lv).toContainText('سطح ۲');
    await expect(lv).toContainText('قهرمان');
    const sk = card.locator(`[data-pp-streak="e2e_${k}_sk"]`);
    await expect(sk).toContainText('روز ۱');
    await expect(card.locator('[data-pp-claims]')).toContainText('جایزهٔ آزمایشی ' + k);
    await sk.locator(`[data-streak-reset="e2e_${k}_sk"]`).click();
    await page.locator('[role="alertdialog"] button', { hasText: 'صفر کردن' }).click();
    await expect(toasts(page).last()).toContainText('صفر شد');
    await expect(sk).toContainText('روز ۰');
    const progress = await eco.get('/users/' + P.id + '/progress?lang=fa');
    expect(progress.streaks.find((x: any) => x.key === 'e2e_' + k + '_sk')).toMatchObject({ count: 0, run: 2 });

    await card.getByRole('link', { name: 'گزارش رویدادها' }).click();
    await expect(page).toHaveURL(new RegExp('#/economy/event-log\\?userId=' + P.id));
    await expect(rows(page).first()).toBeVisible();
    await expect(rows(page).filter({ hasText: 'کنسول' }).first()).toContainText(key);
    await expect(rows(page).filter({ hasText: 'بازی' }).first()).toBeVisible();
    await root(page).locator('.eco-filters select').first().selectOption('admin');
    await expect(rows(page)).toHaveCount(1);
  });

  test('every new page is in the page search, and the section tabs scroll to the current page', async ({ page }) => {
    await open(page, '#/economy/overview');
    await expect(root(page).locator('[data-ov-events]')).toContainText('رویدادهای ۲۴ ساعت اخیر');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('روزهای پیاپی');
    await expect(page.locator('.pl-pal-row').first()).toContainText('روزهای پیاپی');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/streaks/);
    const tabs = page.locator('.pl-side');
    const cur = tabs.locator('a.pl-nav[aria-current="page"]');
    await expect(cur).toContainText('روزهای پیاپی');
    const inView = await cur.evaluate((el) => { const a = el.getBoundingClientRect(); const b = el.parentElement!.getBoundingClientRect(); return a.left >= b.left - 1 && a.right <= b.right + 1; });
    expect(inView).toBe(true);
    await page.locator('.pl-side-search').click();
    await page.locator('.pl-pal-input').fill('جدول سطح');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/level-tables/);
  });
});
