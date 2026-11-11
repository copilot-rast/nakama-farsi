import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, eco, rpc, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { balance, gameRpc, openGame, persian, screen, toast, waitHome, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-ev-${tag}-${k}`);
  players.push(p.id);
  return p;
}

async function adminEvent(userId: string, key: string, value: number) {
  return eco.post('/users/' + userId + '/events', { key, value });
}

test.describe('stargame: events, levels and the daily streak', () => {
  test('a new player collects day one of the daily sign-in reward, and the next day unlocks day two', async ({ page }) => {
    const me = await start(page, 'streak');
    const card = page.locator('#streak-card');
    await expect(card).toBeVisible();
    await expect(card.locator('#streak-title')).toHaveText('جایزهٔ ورود روزانه');
    await expect(card.locator('#streak-info')).toHaveText('۱ روز پشت‌سرهم · بهترین: ۱');
    const day1 = card.locator('[data-day="1"]');
    await expect(day1).toHaveAttribute('data-status', 'claimable');
    await expect(day1).toContainText('۱۰ سکه');
    await expect(card.locator('[data-day="2"]')).toHaveAttribute('data-status', 'locked');
    await expect(page.locator('#rewards-dot')).toBeVisible();
    await expect(page.locator('#balance')).toHaveText('۵۰');
    await day1.click();
    await expect(toast(page)).toContainText('جایزه گرفتی: ۱۰ سکه');
    await expect(page.locator('#balance')).toHaveText('۶۰');
    await expect(card.locator('[data-day="1"]')).toHaveAttribute('data-status', 'claimed');
    await expect(card).toContainText('امروز شمرده شد');
    const again = await gameRpc(page, 'economy_reward_claim', { claimId: (await rpc(me as any, 'economy_rewards_mine', {})).data.items[0].claimId });
    expect(again).toMatchObject({ ok: false, code: 'ALREADY_CLAIMED' });

    sql(`UPDATE economy_user_streak SET last_period = last_period - 1 WHERE user_id='${me.id}'`);
    await page.reload();
    await waitHome(page);
    await expect(card.locator('#streak-info')).toHaveText('۲ روز پشت‌سرهم · بهترین: ۲');
    await expect(card.locator('[data-day="2"]')).toHaveAttribute('data-status', 'claimable');
    await expect(card.locator('[data-day="1"]')).toHaveAttribute('data-status', 'claimed');
    await card.locator('[data-day="2"]').click();
    await expect(page.locator('#balance')).toHaveText('۷۵');

    sql(`UPDATE economy_user_streak SET last_period = last_period - 3 WHERE user_id='${me.id}'`);
    await page.reload();
    await waitHome(page);
    await expect(card.locator('#streak-info')).toHaveText('۱ روز پشت‌سرهم · بهترین: ۲');
    await expect(card).toContainText('جایزهٔ امروزت آماده است');
    await expect(card.locator('[data-day="1"]')).toHaveAttribute('data-status', 'claimable');
  });

  test('experience from rounds raises the level, a level prize waits for a claim, and goals show on the rewards screen', async ({ page }) => {
    const me = await start(page, 'level');
    const level = page.locator('#level-card');
    await expect(level.locator('#level-title')).toHaveText('سطح ۱ · تازه‌کار');
    await expect(level.locator('#level-xp')).toHaveText('۰ از ۱۰۰ امتیاز تجربه');

    const before = await balance(page);
    const out = await adminEvent(me.id, 'round_finished', 120);
    expect(out.levels.find((l: any) => l.key === 'star_levels')).toMatchObject({ to: 2 });
    expect(out.claims.some((c: any) => c.kind === 'level' && c.grantMode === 'claim' && c.status === 'ready')).toBe(true);
    await page.reload();
    await waitHome(page);
    await expect(level.locator('#level-title')).toHaveText('سطح ۲ · ستاره‌چین');
    await expect(level.locator('#level-xp')).toHaveText('۱۲۰ از ۳۰۰ امتیاز تجربه');
    const claim = level.locator('[data-level="2"]');
    await expect(claim).toContainText('جایزهٔ سطح ۲ را بگیر: ۲۰ سکه');
    await claim.click();
    await expect(toast(page)).toContainText('۲۰ سکه');
    await expect(claim).toHaveCount(0);
    await expect.poll(() => balance(page)).toBe(before + 20);

    await page.locator('#menu-rewards').click();
    await screen(page, 'rewards');
    const first = page.locator('[data-rule="star_first_100"]');
    await expect(first).toHaveClass(/ready/);
    await expect(first.locator('[data-progress]')).toHaveText('۱۰۰ از ۱۰۰');
    const rounds = page.locator('[data-rule="star_daily_rounds"]');
    await expect(rounds.locator('[data-progress]')).toContainText('امروز');
    await expect(rounds).toContainText('خودکار می‌رسد');
    await first.locator('[data-claim]').click();
    await expect(toast(page)).toContainText('۵۰ سکه');
    await expect(first.locator('.chip')).toHaveText('انجام شد');
    await expect.poll(() => balance(page)).toBe(before + 70);

    for (let i = 0; i < 10; i++) {
      const r = await gameRpc(page, 'economy_event_record', { key: 'golden_star', value: 1 });
      expect(r.ok).toBe(true);
    }
    const bad = await gameRpc(page, 'economy_event_record', { key: 'round_finished', value: 400 });
    expect(bad).toMatchObject({ ok: false, code: 'EVENT_NOT_ALLOWED' });
    await page.locator('#screen-rewards [data-go="home"]').click();
    await page.locator('#menu-rewards').click();
    const golden = page.locator('[data-rule="star_golden_daily"]');
    await expect(golden).toHaveClass(/ready/);
    await golden.locator('[data-claim]').click();
    await expect.poll(() => balance(page)).toBe(before + 85);

    const after = await adminEvent(me.id, 'round_finished', 0);
    const auto = await adminEvent(me.id, 'round_finished', 0);
    expect([...after.claims, ...auto.claims].some((c: any) => c.status === 'granted' && c.grantMode === 'auto')).toBe(true);
    await page.reload();
    await waitHome(page);
    await expect.poll(() => balance(page)).toBe(before + 105);
    await expect(level.locator('#level-xp')).toContainText(persian(140));
    await page.locator('#menu-rewards').click();
    await expect(rounds.locator('.chip')).toHaveText('انجام شد');
  });

  test('English shows the same progress with Latin digits', async ({ page }) => {
    await openGame(page, `e2e-ev-en-${k}`, { lang: 'en' }).then((p) => players.push(p.id));
    await expect(page.locator('#streak-card [data-day="1"]')).toContainText('Collect');
    await expect(page.locator('#level-card #level-title')).toHaveText('Level 1 · Rookie');
    await page.locator('#menu-rewards').click();
    await expect(page.locator('[data-rule="star_daily_rounds"]')).toContainText('Three rounds a day');
    await expect(page.locator('[data-rule="star_daily_rounds"] [data-progress]')).toHaveText('Today · 0 of 3');
  });
});
