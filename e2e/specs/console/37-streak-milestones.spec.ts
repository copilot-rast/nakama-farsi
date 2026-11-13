import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, devicePlayer, eco, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const skey = 'e2e_' + k + '_ms';
const tkey = 'e2e_' + k + '_visit';
const made = { users: [] as string[], prizes: [] as string[], types: [] as string[], streaks: [] as string[] };
let prize: any;
let P: Player;
let streakId = '';

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

async function playerPage(page: Page, id: string) {
  await page.goto(CONSOLE_URL + '/#/players/' + id);
  await expect(page.locator('[data-player-progress] h2')).toBeVisible({ timeout: 20_000 });
  await expect(page.locator('[data-player-progress] .animate-pulse')).toHaveCount(0);
}

function visit(userId: string) {
  return eco.post('/users/' + userId + '/events', { key: tkey, value: 1 });
}

function userStreak(userId: string, set: string) {
  sql(`UPDATE economy_user_streak SET ${set} WHERE user_id='${userId}' AND streak_id='${streakId}'`);
}

test.beforeAll(async () => {
  prize = await eco.post('/prizes', { key: 'e2e_' + k + '_safe', name: { fa: 'جایزهٔ نقطهٔ امن ' + k, en: 'Safe day prize ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 9 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  const type = await eco.post('/event-types', { key: tkey, name: { fa: 'بازدید ' + k, en: 'Visit ' + k }, clientAllowed: false, isActive: true, rateLimit: 1000 });
  made.types.push(type.id);
  P = await devicePlayer('e2e-ms-' + k, ('ms' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_reward_claim WHERE prize_id='${prize.id}' OR (kind='streak' AND owner_id::text='${streakId || '00000000-0000-0000-0000-000000000000'}'); DELETE FROM economy_user_prize WHERE prize_id='${prize.id}'`);
  for (const id of made.streaks) await eco.del('/streaks/' + id);
  for (const id of made.types) await eco.del('/event-types/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
});

test.describe('streak milestones', () => {
  test('the streak editor saves milestone rows with an optional prize and the re-pay switch, off by default', async ({ page }) => {
    await open(page, '#/economy/streaks/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(skey);
    await fillTranslated(p, 'name', { fa: 'بازدید روزانه ' + k, en: 'Daily visit ' + k });
    await pickIn(field(p, 'رویداد شمارنده'), tkey);
    await field(p, 'طول دوره').locator('input').fill('');
    await expect(p).toContainText('نقطه‌های امن');
    await expect(p.locator('[data-milestones-note]')).toContainText('روز بعدش ۲۱ است');
    await expect(p.locator('[data-rows="milestones"]')).toContainText('نقطهٔ امنی ندارد');
    const repay = p.getByText('جایزهٔ روزها دوباره داده شود');
    await expect(repay).toBeVisible();
    await expect(p).toContainText('جلوی غیبتِ عمدی');

    const ms = p.locator('[data-rows="milestones"]');
    await ms.locator('[data-rows-add="milestones"]').click();
    await ms.locator('[data-row="0"] [data-col="day"] input').fill('۳');
    await pickIn(ms.locator('[data-row="0"] [data-col="prizeId"]'), 'e2e_' + k + '_safe');
    await ms.locator('[data-row="0"] [data-col="grantMode"] select').selectOption('claim');
    await ms.locator('[data-rows-add="milestones"]').click();
    await ms.locator('[data-row="1"] [data-col="day"] input').fill('3');
    await savePanel(page).catch(() => null);
    await expect(ms.locator('[data-row="1"] .eco-items-err')).toContainText('یک نقطهٔ امن');
    await ms.locator('[data-row="1"] [data-col="day"] input').fill('۱۰');
    await savePanel(page);

    const saved = await byKey('streaks', skey);
    streakId = saved.id;
    made.streaks.push(saved.id);
    expect(saved.repayAfterFall).toBe(false);
    expect(saved.cycleLength ?? null).toBe(null);
    expect(saved.milestones.map((m: any) => [m.day, m.prizeId, m.grantMode])).toEqual([[3, prize.id, 'claim'], [10, null, 'claim']]);
    await search(page, skey);
    await expect(rows(page).first().locator('[data-sk-milestones]')).toHaveText('نقطهٔ امن: روز ۳، ۱۰');

    await go(page, '#/economy/streaks/' + saved.id);
    await settlePanel(page);
    await panel(page).getByText('جایزهٔ روزها دوباره داده شود').click();
    await savePanel(page);
    expect((await byKey('streaks', skey)).repayAfterFall).toBe(true);
    await go(page, '#/economy/streaks/' + saved.id);
    await settlePanel(page);
    await panel(page).getByText('جایزهٔ روزها دوباره داده شود').click();
    await savePanel(page);
    const back = await byKey('streaks', skey);
    expect(back.repayAfterFall).toBe(false);
    expect(back.milestones.map((m: any) => m.day)).toEqual([3, 10]);
  });

  test('a miss falls back to the highest milestone, the card shows the safe day, and «صفر کردن» clears it', async ({ page }) => {
    expect(streakId).not.toBe('');
    await visit(P.id);
    userStreak(P.id, 'count=4, peak=4, best=4, last_period=last_period-1');
    const five = await visit(P.id);
    expect(five.streaks.find((s: any) => s.key === skey)).toMatchObject({ count: 5, floor: 3, floored: false, nextMilestone: 10 });
    expect(five.claims.filter((c: any) => c.prizeId === prize.id)).toHaveLength(1);

    userStreak(P.id, 'last_period=last_period-3');
    const fell = await visit(P.id);
    expect(fell.streaks.find((s: any) => s.key === skey)).toMatchObject({ count: 4, floor: 3, floored: true, reset: false });
    expect(fell.claims.filter((c: any) => c.prizeId === prize.id)).toHaveLength(0);

    await playerPage(page, P.id);
    const card = page.locator('[data-player-progress]');
    const sk = card.locator(`[data-pp-streak="${skey}"]`);
    await expect(sk).toContainText('روز ۴ · بهترین: ۵');
    await expect(sk.locator('[data-pp-floor="3"]')).toHaveText('تا روز ۳ امن است · نقطهٔ امن بعدی: روز ۱۰');
    const progress = await eco.get('/users/' + P.id + '/progress?lang=fa');
    const view = progress.streaks.find((x: any) => x.key === skey);
    expect(view).toMatchObject({ floor: 3, nextMilestone: 10, peak: 5, repayAfterFall: false });
    expect(view.milestones.map((m: any) => [m.day, m.reached, m.safe, m.status])).toEqual([[3, true, true, 'claimable'], [10, false, false, 'locked']]);

    await sk.locator(`[data-streak-reset="${skey}"]`).click();
    await page.locator('[role="alertdialog"] button', { hasText: 'صفر کردن' }).click();
    await expect(toasts(page).last()).toContainText('صفر شد');
    await expect(sk).toContainText('روز ۰');
    await expect(sk.locator('[data-pp-floor="0"]')).toHaveText('هنوز نقطهٔ امنی ندارد · نقطهٔ امن بعدی: روز ۳');
    const after = (await eco.get('/users/' + P.id + '/progress?lang=fa')).streaks.find((x: any) => x.key === skey);
    expect(after).toMatchObject({ count: 0, floor: 0, peak: 0, best: 5 });
    expect(Number(sql(`SELECT count(*) FROM economy_reward_claim WHERE user_id='${P.id}' AND scope='m3'`))).toBe(1);
  });

  test('English strings, and the editor at 390px in dark mode without sideways scroll', async ({ page, spawn }) => {
    await open(page, '#/economy/streaks');
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/economy/streaks/' + streakId);
    await settlePanel(page);
    await expect(panel(page)).toContainText('Safe days (milestones)');
    await expect(panel(page)).toContainText('Pay day rewards again');
    await go(page, '#/economy/streaks?q=' + skey);
    await expect(rows(page).first().locator('[data-sk-milestones]')).toHaveText('Safe at day 3, 10');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);

    const m = await spawn({ signedIn: true, mode: 'dark', mobile: true });
    await open(m, '#/economy/streaks/' + streakId);
    await expect(m.locator('html')).toHaveClass(/dark/);
    await settlePanel(m);
    const ms = panel(m).locator('[data-rows="milestones"]');
    await ms.scrollIntoViewIfNeeded();
    await expect(ms.locator('[data-row="1"]')).toBeVisible();
    await noHorizontalScroll(m);
  });
});
