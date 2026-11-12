import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { balance, gameRpc, openGame, screen, toast, waitHome, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-ac-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

async function steer(page: Page) {
  await page.evaluate(() => {
    const w = window as any;
    clearInterval(w.__e2eSteer);
    w.__e2eSteer = setInterval(() => {
      const g = w.__stargame;
      if (g.phase !== 'playing') return;
      const b = g.basket;
      let best: any = null;
      let rank = -1e9;
      for (const o of g.objects) {
        if (o.kind === 'bomb' || o.y > b.y + 4) continue;
        const r = o.y + (o.kind === 'gold' ? 120 : 0) - Math.abs(o.x - b.x) * 0.35;
        if (r > rank) { rank = r; best = o; }
      }
      let x = best ? best.x : b.x;
      for (const o of g.objects) {
        if (o.kind === 'bomb' && o.y > b.y - 110 && o.y < b.y + 12 && Math.abs(o.x - x) < b.w / 2 + 22) x = o.x < g.width / 2 ? o.x + b.w : o.x - b.w;
      }
      const c = document.getElementById('canvas')!;
      const rect = c.getBoundingClientRect();
      c.dispatchEvent(new PointerEvent('pointermove', { clientX: rect.left + x, clientY: rect.top + 10, pointerType: 'mouse', bubbles: true }));
    }, 25);
  });
}

async function openAch(page: Page) {
  await page.locator('#menu-achievements').click();
  await screen(page, 'achievements');
  await expect(page.locator('#achievements-list .spinner')).toHaveCount(0);
}

async function home(page: Page) {
  await page.locator('#screen-achievements [data-go="home"]').click();
  await waitHome(page);
}

test.describe('stargame: achievements', () => {
  test('rounds unlock the first badge and the bronze tier, its prize is collected once, and the console sees it', async ({ page, spawn }) => {
    test.setTimeout(150_000);
    const me = await start(page, 'play');
    await openAch(page);
    await expect(page.locator('#achievements-summary')).toHaveText('۰ از ۵ دستاورد · ۰ امتیاز');
    const first = page.locator('[data-achievement="star_ach_first"]');
    await expect(first).toHaveAttribute('data-state', 'locked');
    const rounds = page.locator('[data-achievement="star_ach_rounds"]');
    await expect(rounds.locator('[data-ach-tier]')).toHaveText('ردهٔ ۰ از ۳');
    await expect(rounds.locator('[data-progress]')).toHaveText('تا برنزی: ۰ از ۳');
    await expect(rounds.locator('[data-tier="1"]')).toContainText('برنزی');
    const secret = page.locator('[data-achievement="hidden"]');
    await expect(secret).toHaveCount(1);
    await expect(secret).toContainText('؟؟؟');
    await expect(secret.locator('[data-ach-hint]')).toHaveText('یک دور واقعاً بزرگ بازی کن.');
    await expect(page.locator('[data-achievement="star_ach_legend"]')).toHaveCount(0);
    await expect(page.locator('[data-achievement="star_ach_hunter"]')).toContainText('شکارچی کامل');

    await home(page);
    await page.locator('#btn-play').click();
    await screen(page, 'play');
    await page.waitForFunction(() => (window as any).__stargame.phase === 'playing', null, { timeout: 15_000 });
    await steer(page);
    await page.waitForFunction(() => (window as any).__stargame.phase === 'done', null, { timeout: 60_000 });
    const overlay = page.locator('#stage-overlay');
    await expect(overlay.locator('[data-round-notes]')).toContainText('دستاورد تازه باز شد');
    await overlay.getByRole('button', { name: 'خانه' }).click();
    await waitHome(page);

    await eco.post('/users/' + me.id + '/events', { key: 'round_finished', value: 10 });
    await eco.post('/users/' + me.id + '/events', { key: 'round_finished', value: 10 });
    await page.reload();
    await waitHome(page);
    await expect(page.locator('#achievements-dot')).toBeVisible();
    await openAch(page);
    await expect(first).toHaveAttribute('data-state', 'completed');
    await expect(first.locator('[data-ach-state]')).toHaveText('همهٔ رده‌ها');
    await expect(rounds).toHaveAttribute('data-tier', '1');
    await expect(rounds.locator('[data-ach-tier]')).toHaveText('ردهٔ ۱ از ۳، برنزی');
    await expect(rounds.locator('[data-progress]')).toHaveText('تا نقره‌ای: ۳ از ۱۰');
    await expect(rounds.locator('[data-tier="1"]')).toHaveAttribute('data-tier-status', 'claimable');
    const before = await balance(page);
    await rounds.locator('[data-ach-claim]').click();
    await expect(toast(page)).toContainText('۲۰ سکه');
    await expect.poll(() => balance(page)).toBe(before + 20);
    await expect(rounds.locator('[data-tier="1"]')).toHaveAttribute('data-tier-status', 'claimed');
    await expect(rounds.locator('[data-ach-claim]')).toHaveCount(0);
    const again = await gameRpc(page, 'economy_achievement_claim', { key: 'star_ach_rounds' });
    expect(again).toMatchObject({ ok: false, code: 'ALREADY_CLAIMED' });
    const ledger = sql(`SELECT count(*) FROM wallet_ledger WHERE user_id='${me.id}' AND metadata->>'sourceType'='achievement'`);
    expect(Number(ledger)).toBe(1);
    const unlocks = sql(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='achievement_unlocked' AND e.user_id='${me.id}'`);
    expect(Number(unlocks)).toBeGreaterThanOrEqual(2);

    const c = await spawn({ signedIn: true });
    await c.goto(CONSOLE_URL + '/#/players/' + me.id);
    const card = c.locator('[data-player-achievements]');
    await expect(card.locator('h2')).toHaveText('دستاوردهای بازیکن', { timeout: 20_000 });
    await expect(card.locator('[data-pac="star_ach_rounds"] [data-pac-tier]')).toHaveText('ردهٔ ۱ از ۳');
    await expect(card.locator('[data-pac="star_ach_rounds"] [data-pac-claim]')).toHaveCount(0);
    await expect(card.locator('[data-pac="star_ach_first"]')).toHaveAttribute('data-state', 'completed');
    const claims = await eco.get('/event-claims?userId=' + me.id + '&kind=achievement');
    expect(claims.items.map((x: any) => x.sourceKey)).toEqual(['star_ach_rounds']);
  });

  test('the STARFAN gift unlocks the fan badge with an in-app note, and a friend sees it on the profile', async ({ page }) => {
    const me = await start(page, 'fan');
    await page.locator('#screen-home [data-go="gift"]').click();
    await screen(page, 'gift');
    await page.locator('#gift-code').fill('starfan');
    await page.locator('#gift-form button[type="submit"]').click();
    await expect(page.locator('#gift-msg')).toHaveClass(/ok/);
    await page.locator('#screen-gift [data-go="home"]').click();
    await waitHome(page);
    await openAch(page);
    const fan = page.locator('[data-achievement="star_ach_supporter"]');
    await expect(fan).toHaveAttribute('data-state', 'completed');
    await expect(page.locator('#achievements-summary')).toHaveText('۱ از ۵ دستاورد · ۱۵ امتیاز');
    const note = sql(`SELECT n.title->>'fa' || '|' || (n.action->>'value') FROM publication_notification n JOIN publication_notification_recipient r ON r.notification_id=n.id WHERE r.user_id='${me.id}' AND n.created_by='economy:achievement'`);
    expect(note).toBe('دستاورد تازه|achievements');
    const again = await gameRpc(page, 'economy_gift_redeem', { code: 'STARFAN' });
    expect(again.ok).toBe(false);
    const prof = await gameRpc(page, 'economy_achievements_profile', { userId: me.id, lang: 'fa' });
    expect(prof.ok).toBe(true);
    expect((prof as any).data).toMatchObject({ unlocked: 1, points: 15 });
    expect((prof as any).data.items.map((x: any) => x.key)).toEqual(['star_ach_supporter']);
  });

  test('English shows achievements with Latin digits at 390px', async ({ page }) => {
    await start(page, 'en', { lang: 'en' });
    await expect(page.locator('#menu-achievements')).toContainText('Achievements');
    await page.setViewportSize({ width: 390, height: 844 });
    await openAch(page);
    await expect(page.locator('#achievements-summary')).toHaveText('0 of 5 achievements · 0 points');
    const rounds = page.locator('[data-achievement="star_ach_rounds"]');
    await expect(rounds).toContainText('Busy star catcher');
    await expect(rounds.locator('[data-progress]')).toHaveText('To Bronze: 0 of 3');
    await expect(page.locator('[data-achievement="hidden"]')).toContainText('Secret achievement');
    await noHorizontalScroll(page);
  });
});
