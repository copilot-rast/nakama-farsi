import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, eco, rpc, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { balance, gameRpc, openGame, persian, screen, toast, waitHome, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-tk-${tag}-${k}`, opts);
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

async function openTasks(page: Page) {
  await page.locator('#menu-tasks').click();
  await screen(page, 'tasks');
  await expect(page.locator('#tasks-list .spinner')).toHaveCount(0);
}

test.describe('stargame: missions', () => {
  test('a played round moves the missions, a finished mission pays its coins once, and the console sees it', async ({ page, spawn }) => {
    test.setTimeout(150_000);
    const me = await start(page, 'play');
    await openTasks(page);
    const two = page.locator('[data-task="star_mission_two_rounds"]');
    await expect(two).toHaveAttribute('data-state', 'active');
    await expect(two.locator('[data-progress]')).toHaveText('امروز، ۰ از ۲');
    await expect(two).toContainText('۲۰ سکه');
    const chain = page.locator('[data-group="star_hunter_chain"]');
    await expect(chain).toContainText('زنجیرهٔ شکارچی');
    await expect(chain.locator('[data-group-progress]')).toHaveText('۰ از ۳ مأموریت');
    await expect(chain.locator('[data-task="star_chain_one"]')).toBeVisible();
    await expect(chain.locator('[data-task="star_chain_two"]')).toHaveCount(0);
    const starter = page.locator('[data-task="star_mission_starter"]');
    await expect(starter.locator('[data-progress]').first()).toHaveText('۰ از ۳ قدم');
    await expect(page.locator('[data-task="star_mission_veteran"]')).toHaveCount(0);

    await page.locator('#screen-tasks [data-go="home"]').click();
    await waitHome(page);
    await page.locator('#btn-play').click();
    await screen(page, 'play');
    await page.waitForFunction(() => (window as any).__stargame.phase === 'playing', null, { timeout: 15_000 });
    await steer(page);
    await page.waitForFunction(() => (window as any).__stargame.phase === 'done', null, { timeout: 60_000 });
    const overlay = page.locator('#stage-overlay');
    await expect(overlay).toContainText('دور تمام شد!');
    await expect(overlay.locator('[data-round-notes]')).toContainText('مأموریت تمام شد');
    await overlay.getByRole('button', { name: 'خانه' }).click();
    await waitHome(page);
    await expect(page.locator('#tasks-dot')).toBeVisible();

    await openTasks(page);
    await expect(two.locator('[data-progress]')).toHaveText('امروز، ۱ از ۲');
    const step = starter.locator('[data-task="star_mission_starter_round"]');
    await expect(step).toHaveAttribute('data-state', 'claimed');
    await expect(step.locator('[data-task-state]')).toHaveText('تمام شد');
    const one = chain.locator('[data-task="star_chain_one"]');
    await expect(one).toHaveAttribute('data-state', 'completed');
    const before = await balance(page);
    await one.locator('[data-claim]').click();
    await expect(toast(page)).toContainText('جایزهٔ مأموریت رسید: ۱۰ سکه');
    await expect.poll(() => balance(page)).toBe(before + 10);
    await expect(chain.locator('[data-group-progress]')).toHaveText('۱ از ۳ مأموریت');
    await expect(chain.locator('[data-task="star_chain_two"]')).toHaveAttribute('data-state', 'active');
    const again = await gameRpc(page, 'economy_task_claim', { key: 'star_chain_one' });
    expect(again).toMatchObject({ ok: false, code: 'ALREADY_CLAIMED' });

    await eco.post('/users/' + me.id + '/events', { key: 'round_finished', value: 10 });
    await page.locator('#screen-tasks [data-go="home"]').click();
    await waitHome(page);
    await openTasks(page);
    await expect(two).toHaveAttribute('data-state', 'completed');
    await two.locator('[data-claim]').click();
    await expect(toast(page)).toContainText('۲۰ سکه');
    await expect.poll(() => balance(page)).toBe(before + 30);
    await expect(two).toHaveAttribute('data-state', 'claimed');
    await expect(two).toContainText('جایزه را گرفتی');

    const wallet = await rpc(me as any, 'economy_balances', {});
    expect(wallet.data.wallet.coin).toBe(before + 30);
    const ledger = sql(`SELECT count(*) FROM wallet_ledger WHERE user_id='${me.id}' AND metadata->>'sourceType'='task'`);
    expect(Number(ledger)).toBe(2);
    const claims = await eco.get('/event-claims?userId=' + me.id + '&kind=task');
    expect(claims.items.map((c: any) => c.sourceKey).sort()).toEqual(['star_chain_one', 'star_mission_two_rounds']);

    const c = await spawn({ signedIn: true });
    await c.goto(CONSOLE_URL + '/#/players/' + me.id);
    const card = c.locator('[data-player-tasks]');
    await expect(card.locator('h2')).toHaveText('تسک‌های بازیکن', { timeout: 20_000 });
    await expect(card.locator('[data-pp-task="star_mission_two_rounds"]')).toHaveAttribute('data-state', 'claimed');
    await expect(card.locator('[data-pp-group="star_hunter_chain"] [data-pp-task="star_chain_two"]')).toHaveAttribute('data-state', 'active');
    const recent = c.locator('[data-player-progress] [data-pp-claims]');
    await expect(recent).toContainText('شکار اول: یک دور بازی کن');
  });

  test('the starter mission pays after its three steps and opens the veteran mission at level 2 and sends an in-app note', async ({ page }) => {
    const me = await start(page, 'starter');
    for (const [key, value] of [['round_finished', 120], ['golden_star', 1], ['golden_star', 1], ['golden_star', 1]] as [string, number][]) {
      await eco.post('/users/' + me.id + '/events', { key, value });
    }
    await page.reload();
    await waitHome(page);
    await openTasks(page);
    const starter = page.locator('[data-task="star_mission_starter"]');
    await expect(starter).toHaveAttribute('data-state', 'completed');
    await expect(starter.locator('[data-progress]').first()).toHaveText('۳ از ۳ قدم');
    const before = await balance(page);
    await starter.locator('[data-claim]').first().click();
    await expect(toast(page)).toContainText('۲۰۰ سکه');
    await expect.poll(() => balance(page)).toBe(before + 200);
    const veteran = page.locator('[data-task="star_mission_veteran"]');
    await expect(veteran).toHaveAttribute('data-state', 'active');
    await expect(veteran.locator('[data-progress]')).toHaveText(persian('0') + ' از ' + persian(10));
    const note = sql(`SELECT n.title->>'fa' || '|' || (n.action->>'value') FROM publication_notification n JOIN publication_notification_recipient r ON r.notification_id=n.id WHERE r.user_id='${me.id}' AND n.created_by='economy:task'`);
    expect(note).toBe('مأموریت انجام شد|tasks');

    await page.setViewportSize({ width: 390, height: 844 });
    await noHorizontalScroll(page);
    await page.setViewportSize({ width: 1280, height: 900 });
    const en = await gameRpc(page, 'economy_tasks_list', { lang: 'en' });
    expect(en.ok).toBe(true);
    expect((en as any).data.items.find((x: any) => x.key === 'star_mission_veteran').name).toBe('Veteran: play ten rounds');
  });

  test('English shows the missions with Latin digits', async ({ page }) => {
    await start(page, 'en', { lang: 'en' });
    await expect(page.locator('#menu-tasks')).toContainText('Missions');
    await openTasks(page);
    const two = page.locator('[data-task="star_mission_two_rounds"]');
    await expect(two).toContainText('Play two rounds today');
    await expect(two.locator('[data-progress]')).toHaveText('Today · 0 of 2');
    await expect(page.locator('[data-group="star_hunter_chain"] [data-group-progress]')).toHaveText('0 of 3 missions');
    await page.setViewportSize({ width: 390, height: 844 });
    await noHorizontalScroll(page);
  });

  test('a client cannot push a mission with events the server keeps for itself', async ({ page }) => {
    await start(page, 'cheat');
    const bad = await gameRpc(page, 'economy_event_record', { key: 'round_finished', value: 400 });
    expect(bad).toMatchObject({ ok: false, code: 'EVENT_NOT_ALLOWED' });
    const big = await gameRpc(page, 'economy_event_record', { key: 'golden_star', value: 50 });
    expect(big).toMatchObject({ ok: false, code: 'VALUE_OUT_OF_RANGE' });
    const list = await gameRpc(page, 'economy_tasks_list', { lang: 'fa' });
    const two = (list as any).data.items.find((x: any) => x.key === 'star_mission_two_rounds');
    expect(two.value).toBe(0);
  });
});
