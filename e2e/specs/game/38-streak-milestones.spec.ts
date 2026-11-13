import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { balance, gameRpc, openGame, toast, waitHome, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, lang = 'fa'): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-ms-${tag}-${k}`, { lang });
  players.push(p.id);
  return p;
}

function setStreak(userId: string, set: string) {
  sql(`UPDATE economy_user_streak u SET ${set} FROM economy_streak s WHERE s.id=u.streak_id AND s.key='daily_login' AND u.user_id='${userId}'`);
}

test.describe('stargame: the safe day of the daily streak', () => {
  test('the seeded day 7 milestone: the next safe day, its prize, and a miss that falls back to day 7 instead of day one', async ({ page }) => {
    const sk = (await eco.get('/streaks?q=daily_login&limit=5')).items.find((x: any) => x.key === 'daily_login');
    expect(sk.milestones.map((m: any) => m.day)).toEqual([7]);
    expect(sk.repayAfterFall).toBe(false);

    const me = await start(page, 'safe');
    const card = page.locator('#streak-card');
    await expect(card.locator('#streak-info')).toHaveText('۱ روز پشت‌سرهم · بهترین: ۱');
    await expect(card.locator('#streak-floor')).toHaveCount(0);
    await expect(card.locator('#streak-next')).toHaveText('نقطهٔ امن بعدی: روز ۷ (۶ روز دیگر)');
    await expect(card.locator('[data-milestone="7"]')).toHaveAttribute('data-status', 'locked');
    await expect(card.locator('[data-milestone="7"]')).toContainText('روز ۷: ۱۰۰ سکه');

    setStreak(me.id, 'count=7, peak=7, best=7, last_period=u.last_period-1');
    await page.reload();
    await waitHome(page);
    await expect(card.locator('#streak-info')).toHaveText('۸ روز پشت‌سرهم · بهترین: ۸');
    await expect(card.locator('#streak-floor')).toHaveText('تا روز ۷ امن است');
    await expect(card.locator('#streak-next')).toHaveCount(0);
    const chip = card.locator('[data-milestone="7"]');
    await expect(chip).toHaveAttribute('data-status', 'claimable');
    const before = await balance(page);
    await chip.click();
    await expect(toast(page)).toContainText('۱۰۰ سکه');
    await expect.poll(() => balance(page)).toBe(before + 100);
    await expect(card.locator('[data-milestone="7"]')).toHaveAttribute('data-status', 'claimed');

    setStreak(me.id, 'last_period=u.last_period-3');
    const broken = await gameRpc(page, 'economy_streaks', {});
    const view = broken.data.items.find((x: any) => x.key === 'daily_login');
    expect(view).toMatchObject({ broken: true, floored: true, count: 7, floor: 7 });
    await page.evaluate(() => (window as any).eval('refreshProgress()'));
    await expect(card).toContainText('به روز ۷ برگشتی، نه روز اول');

    await page.reload();
    await waitHome(page);
    await expect(card.locator('#streak-info')).toHaveText('۸ روز پشت‌سرهم · بهترین: ۸');
    await expect(card.locator('#streak-floor')).toHaveText('تا روز ۷ امن است');
    expect(sql(`SELECT u.run || ':' || u.falls FROM economy_user_streak u JOIN economy_streak s ON s.id=u.streak_id WHERE s.key='daily_login' AND u.user_id='${me.id}'`)).toBe('1:1');
    expect(sql(`SELECT count(*) FROM economy_reward_claim c JOIN economy_streak s ON s.id=c.owner_id WHERE s.key='daily_login' AND c.user_id='${me.id}' AND c.scope='m7'`)).toBe('1');
  });

  test('English at 390px shows the next safe day without sideways scroll', async ({ spawn }) => {
    const m = await spawn({ mobile: true });
    await start(m, 'en', 'en');
    const card = m.locator('#streak-card');
    await expect(card.locator('#streak-next')).toHaveText('Next safe day: day 7 (6 more days)');
    await expect(card.locator('[data-milestone="7"]')).toContainText('Day 7: 100 coins');
    await noHorizontalScroll(m);
  });
});
