import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { balance, gameRpc, openGame, screen, waitHome, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];
const boxKey = 'e2e_' + k + '_box';

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  sql(`DELETE FROM economy_chance_play WHERE game_id IN (SELECT id FROM economy_chance_game WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/chance-games?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/chance-games/' + it.id)).status).toBe(200);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-chg-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

async function openChance(page: Page) {
  await page.locator('#menu-chance').click();
  await screen(page, 'chance');
  await expect(page.locator('#chance-list .spinner')).toHaveCount(0);
  await expect(page.locator('[data-chance-game]').first()).toBeVisible();
}

function slotOf(rows: any[], roll: number) {
  let cum = 0;
  for (let i = 0; i < rows.length; i++) {
    cum += Math.round(Number(rows[i].chance) * 10000);
    if (roll * 10000 < cum) return i;
  }
  return -1;
}

test.describe('stargame: chance games', () => {
  test('the daily wheel spins once, lands on the slot the server picked, pays it and waits for the next turn', async ({ page, spawn }) => {
    const me = await start(page, 'free');
    await expect(page.locator('#chance-dot')).toHaveText('۱');
    await openChance(page);
    const wheel = page.locator('[data-chance-game="star_wheel_daily"]');
    await expect(wheel).toContainText('گردونهٔ روزانه');
    await expect(wheel).toContainText('رایگان، هر ۱ روز یک بار');
    await expect(wheel.locator('[data-chance-status]')).toHaveText('نوبت رایگانت آماده است');
    await expect(wheel.locator('svg [data-slice]')).toHaveCount(4);
    const box = page.locator('[data-chance-game="star_lucky_box"]');
    await expect(box.locator('[data-chance-ticket="star_lucky_box"]')).toHaveText('با بلیت · ۱۵ سکه');
    await expect(box.locator('[data-chance-status]')).toHaveText('این بازی نوبت رایگان ندارد');
    await expect(page.locator('[data-chance-free="star_lucky_box"]')).toHaveCount(0);
    expect(await balance(page)).toBe(50);

    await wheel.locator('[data-chance-free="star_wheel_daily"]').click();
    const result = page.locator('[data-chance-game="star_wheel_daily"] [data-chance-result]');
    await expect(result).toHaveClass(/ch-(won|lost)/, { timeout: 10_000 });
    const hist = await gameRpc(page, 'economy_chance_history', { lang: 'fa' });
    expect(hist.ok).toBe(true);
    const play = (hist as any).data.items[0];
    expect(play).toMatchObject({ gameKey: 'star_wheel_daily', mode: 'free', status: 'done' });
    await expect(result).toHaveAttribute('data-prize-id', play.prizeId || '');
    const games = await gameRpc(page, 'economy_chance_games', { lang: 'fa' });
    const daily = (games as any).data.items.find((g: any) => g.key === 'star_wheel_daily');
    expect(slotOf(daily.prize, play.roll)).toBe(play.slot);
    const rot = Number(await page.locator('[data-chance-game="star_wheel_daily"] [data-chance-wheel]').getAttribute('data-rotation'));
    const landed = ((360 - (rot % 360)) % 360) / 3.6;
    expect(Math.abs(landed - play.roll)).toBeLessThan(0.01);
    const coins = play.prizeId ? Number(play.rewards[0].value) : 0;
    if (play.prizeId) await expect(result).toContainText('بردی: ' + play.prize.name);
    else await expect(result).toHaveText('این بار پوچ شد. دفعهٔ بعد!');
    await expect.poll(() => balance(page)).toBe(50 + coins);
    await expect(page.locator('[data-chance-game="star_wheel_daily"] [data-chance-free]')).toBeDisabled();
    await expect(page.locator('[data-chance-game="star_wheel_daily"] [data-chance-wait]')).toContainText('نوبت رایگان بعدی: ۲۳:۵');
    await expect(page.locator('#chance-history [data-chance-play]')).toHaveCount(1);
    await expect(page.locator('#chance-history [data-chance-play]').first()).toContainText('گردونهٔ روزانه');

    const again = await gameRpc(page, 'economy_chance_play', { id: daily.id });
    expect(again).toMatchObject({ ok: false, code: 'WAITING_FOR_TURN', status: 403 });
    const events = sql(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='ChanceGame:Played' AND e.user_id='${me.id}' AND e.metadata->>'entityId'='${daily.id}'`);
    expect(Number(events)).toBe(1);
    const claims = sql(`SELECT count(*) || '|' || count(*) FILTER (WHERE status='granted') FROM economy_reward_claim WHERE kind='chance' AND user_id='${me.id}'`);
    expect(claims).toBe(play.prizeId ? '1|1' : '0|0');

    const admin = await spawn({ signedIn: true });
    await admin.goto(CONSOLE_URL + '/#/players/' + me.id);
    const card = admin.locator('[data-player-chance]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('[data-pch]')).toHaveCount(1);
    await expect(card.locator('[data-pch]').first()).toContainText(play.prizeId ? 'برد: ' + play.prize.name : 'پوچ');

    sql(`UPDATE economy_chance_play SET created_at = created_at - interval '25 hours' WHERE user_id='${me.id}'`);
    await page.locator('#screen-chance [data-go="home"]').click();
    await waitHome(page);
    await expect(page.locator('#chance-dot')).toHaveText('۱');
    await openChance(page);
    await expect(page.locator('[data-chance-game="star_wheel_daily"] [data-chance-free]')).toBeEnabled();
  });

  test('ticket plays take the price every time, show nothing won, and stop when the coins run out', async ({ page }) => {
    const coin = (await eco.get('/currencies?q=coin&limit=20')).items.find((c: any) => c.key === 'coin');
    await eco.post('/chance-games', { key: boxKey, name: { fa: 'جعبهٔ خالی ' + k, en: 'Empty box ' + k }, isActive: true, sort: 9, prizeType: 'prize', price: { currencyId: coin.id, amount: 20 }, prize: [{ chance: 100 }] });
    const me = await start(page, 'ticket');
    await openChance(page);
    const card = page.locator(`[data-chance-game="${boxKey}"]`);
    const btn = card.locator(`[data-chance-ticket="${boxKey}"]`);
    await expect(btn).toHaveText('با بلیت · ۲۰ سکه');
    await btn.click();
    await expect(card.locator('[data-chance-result]')).toHaveText('این بار پوچ شد. دفعهٔ بعد!', { timeout: 10_000 });
    await expect.poll(() => balance(page)).toBe(30);
    await page.locator(`[data-chance-game="${boxKey}"] [data-chance-ticket]`).click();
    await expect.poll(() => balance(page)).toBe(10);
    await expect(page.locator(`[data-chance-game="${boxKey}"] [data-chance-ticket]`)).toBeDisabled();
    await expect(page.locator(`[data-chance-game="${boxKey}"] [data-chance-short]`)).toHaveText('برای بلیت ۲۰ سکه لازم است.');
    await expect(page.locator(`[data-chance-game="${boxKey}"] [data-chance-plays]`)).toHaveText('۲ بار بازی کرده‌ای');
    const game = (await eco.get('/chance-games?q=' + boxKey)).items[0];
    const short = await gameRpc(page, 'economy_chance_play_with_ticket', { id: game.id });
    expect(short).toMatchObject({ ok: false, code: 'CURRENCY_IS_NOT_ENOUGH', status: 400 });
    const debits = sql(`SELECT count(*) FROM wallet_ledger WHERE user_id='${me.id}' AND metadata->>'sourceType'='chance' AND metadata->>'part'='debit'`);
    expect(Number(debits)).toBe(2);
    const rows = page.locator('#chance-history [data-chance-play]');
    await expect(rows).toHaveCount(2);
    await expect(rows.first()).toContainText('بلیت');
    await expect(rows.first()).toContainText('پوچ');
    expect(await balance(page)).toBe(10);
  });

  test('English shows the wheel with Latin digits at 390px', async ({ page }) => {
    await start(page, 'en', { lang: 'en' });
    await expect(page.locator('#menu-chance')).toContainText('Luck');
    await page.setViewportSize({ width: 390, height: 844 });
    await openChance(page);
    const wheel = page.locator('[data-chance-game="star_wheel_daily"]');
    await expect(wheel).toContainText('Daily wheel');
    await expect(wheel.locator('[data-chance-free]')).toHaveText('Spin free');
    await expect(page.locator('[data-chance-ticket="star_lucky_box"]')).toHaveText('Ticket · 15 Coin');
    await expect(wheel).toContainText('Free once every 1 d');
    await noHorizontalScroll(page);
  });
});
