import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { consoleApi, deleteUser, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { balance, gameRpc, openGame, persian, screen, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];
const restore: Record<string, any> = {};

test.describe.configure({ mode: 'serial' });

async function gameRow(key: string) {
  const res = await eco.get('/games?q=' + key + '&limit=20');
  return (res.items || []).find((x: any) => x.key === key);
}

async function putGame(key: string, change: Record<string, any>) {
  const g = await gameRow(key);
  if (!restore[key]) restore[key] = JSON.parse(JSON.stringify(g));
  const res = await consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, change));
  expect(res.status).toBe(200);
}

async function putBack(key: string) {
  const g = restore[key];
  if (!g) return;
  const res = await consoleApi('PUT', '/v2/console/economy/games/' + g.id, g);
  expect(res.status).toBe(200);
  delete restore[key];
}

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  for (const key of Object.keys(restore)) await putBack(key);
  sql(`DELETE FROM economy_game_match m WHERE NOT EXISTS (SELECT 1 FROM economy_game_entry e WHERE e.match_id=m.id AND EXISTS (SELECT 1 FROM users u WHERE u.id=e.user_id))`);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-gm-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

async function openLobby(page: Page) {
  await page.locator('#menu-games').click();
  await screen(page, 'games');
  await expect(page.locator('#games-list .spinner')).toHaveCount(0);
}

test.describe('games: Star Catcher lobby', () => {
  test('lobby lists Star Catcher and tic-tac-toe with rooms, fees, prizes and level locks', async ({ page }) => {
    await start(page, 'lobby');
    await openLobby(page);
    const star = page.locator('[data-game="stargame"]');
    const ttt = page.locator('[data-game="stargame_tictactoe"]');
    await expect(star).toContainText('ستاره‌چین');
    await expect(star).toContainText('تک‌نفره');
    await expect(star).toContainText('هر دور ۳۰ ثانیه');
    await expect(star).toContainText('آرکید');
    await expect(ttt).toContainText('دوز');
    await expect(ttt).toContainText('۲ نفره');
    await expect(ttt).toContainText('هر نوبت ۱۵ ثانیه');
    await expect(ttt.locator('[data-room-card]')).toHaveCount(3);
    await expect(ttt.locator('[data-room-card="bronze"]')).toContainText('ورودی ۱۰ سکه');
    await expect(ttt.locator('[data-room-card="bronze"]')).toContainText('جایزه: ۲۰ سکه');
    await expect(ttt.locator('[data-room-card="silver"]')).toContainText('از سطح ۲');
    await expect(ttt.locator('[data-room-card="silver"] button')).toBeDisabled();
    await expect(ttt.locator('[data-room-card="gold"]')).toContainText('جایزه: گنج بزرگ');
    await expect(ttt.locator('[data-room-card="bronze"] button')).toBeEnabled();
    const order = await page.locator('#games-list [data-game]').evaluateAll((els) => els.map((e) => e.getAttribute('data-game')));
    expect(order.slice(0, 2)).toEqual(['stargame', 'stargame_tictactoe']);
  });

  test('Star Catcher reads its round time from the game settings', async ({ page }) => {
    await start(page, 'round');
    await putGame('stargame', { settings: { roundSeconds: 25, maxCoins: 12 } });
    const r = await gameRpc(page, 'stargame_start', {});
    expect(r.ok).toBe(true);
    expect(r.data).toMatchObject({ seconds: 25, maxCoins: 12 });
    await openLobby(page);
    await expect(page.locator('[data-game="stargame"]')).toContainText('هر دور ۲۵ ثانیه');
    await page.locator('[data-game="stargame"] [data-play]').click();
    await screen(page, 'play');
    await page.waitForFunction(() => (window as any).__stargame.phase === 'countdown' || (window as any).__stargame.phase === 'playing', null, { timeout: 15_000 });
    await expect(page.locator('#hud-time')).toHaveText('۲۵');
    expect(await page.evaluate(() => (window as any).eval('game.roundSeconds'))).toBe(25);
    await putGame('stargame', { isActive: false });
    const off = await gameRpc(page, 'stargame_start', {});
    expect(off).toMatchObject({ ok: false, code: 'GAME_DISABLED' });
    await putBack('stargame');
    const back = await gameRpc(page, 'stargame_start', {});
    expect(back.data.seconds).toBe(30);
  });

  test('two players match in the bronze room, pay the entry once, the winner gets the prize', async ({ page, spawn }) => {
    const a = await start(page, 'ttt-a');
    const other = await spawn({ signedIn: false });
    const b = await start(other, 'ttt-b');
    for (const pg of [page, other]) {
      await openLobby(pg);
      await pg.locator('[data-game="stargame_tictactoe"] [data-room-card="bronze"] button').click();
      await expect(pg.locator('#games-search')).toContainText('دنبال حریف می‌گردیم');
    }
    for (const pg of [page, other]) await expect(pg.locator('#screen-ttt')).toBeVisible({ timeout: 60_000 });
    for (const pg of [page, other]) await expect(pg.locator('#ttt-board [data-cell]')).toHaveCount(9, { timeout: 45_000 });
    const marks = await page.evaluate(() => (window as any).__lobby.view.marks);
    const x = marks[a.id] === 'X' ? page : other;
    const o = x === page ? other : page;
    const xId = x === page ? a.id : b.id;
    const matchId = await page.evaluate(() => (window as any).__lobby.matchId);
    await expect(x.locator('#ttt-board')).toHaveAttribute('data-turn', 'me');
    await expect(x.locator('#ttt-status')).toContainText('نوبت توست');
    await expect(o.locator('#ttt-status')).toContainText('نوبت حریف است');
    await expect(x.locator('#ttt-status')).toContainText('تو: X');
    const entries = Number(sql(`SELECT count(*) FROM economy_game_entry WHERE match_id='${matchId}' AND status='paid'`));
    expect(entries).toBe(2);
    for (const [pg, cell] of [[x, 0], [o, 3], [x, 1], [o, 4], [x, 2]] as [Page, number][]) {
      await expect(pg.locator('#ttt-board')).toHaveAttribute('data-turn', 'me');
      await pg.locator(`#ttt-board [data-cell="${cell}"]`).click();
      await expect(pg.locator(`#ttt-board [data-cell="${cell}"]`)).not.toHaveText('');
    }
    await expect(x.locator('#ttt-result')).toHaveAttribute('data-outcome', 'won');
    await expect(x.locator('#ttt-result')).toContainText('بردی!');
    await expect(x.locator('#ttt-result')).toContainText('سه خانه در یک خط');
    await expect(x.locator('#ttt-result .ttt-prize')).toContainText('جایزه‌ات رسید: ۲۰ سکه');
    await expect(o.locator('#ttt-result')).toHaveAttribute('data-outcome', 'lost');
    await expect(o.locator('#ttt-result')).toContainText('باختی.');
    await expect.poll(async () => balance(x)).toBe(60);
    await expect.poll(async () => balance(o)).toBe(40);
    expect(sql(`SELECT status || ':' || array_to_string(winners, ',') FROM economy_game_match WHERE id='${matchId}'`)).toBe('finished:' + xId);
    expect(sql(`SELECT count(*) FROM wallet_ledger WHERE metadata->>'sourceId'='${matchId}' AND metadata->>'reason'='game_entry'`)).toBe('2');
    expect(sql(`SELECT count(*) FROM economy_user_prize WHERE source_type='game' AND user_id='${xId}' AND claimed`)).toBe('1');
    expect(sql(`SELECT score FROM leaderboard_record WHERE leaderboard_id='stargame_tictactoe_wins' AND owner_id='${xId}'`)).toBe('1');
    const detail = await consoleApi('GET', '/v2/console/economy/game-matches/' + encodeURIComponent(matchId));
    expect(detail.data).toMatchObject({ status: 'finished', gameKey: 'stargame_tictactoe', roomKey: 'bronze', players: 2, fees: 20 });
    expect(detail.data.entries.map((e: any) => e.prizeStatus).sort()).toEqual(['none', 'paid']);
    await x.locator('#ttt-back').click();
    await screen(x, 'games');
  });

  test('a switched-off game shows its maintenance message and cannot be entered', async ({ page }) => {
    await start(page, 'maint');
    await putGame('stargame_tictactoe', { isActive: false });
    await openLobby(page);
    const ttt = page.locator('[data-game="stargame_tictactoe"]');
    await expect(ttt).toHaveAttribute('data-active', 'false');
    await expect(ttt.locator('.game-maint')).toHaveText('دوز فعلاً بسته است. کمی بعد سر بزن.');
    await expect(ttt).toContainText('در حال تعمیر');
    for (const b of await ttt.locator('[data-room-card] button').all()) await expect(b).toBeDisabled();
    const join = await gameRpc(page, 'economy_game_join', { key: 'stargame_tictactoe', room: 'bronze', platform: 'web' });
    expect(join).toMatchObject({ ok: false, code: 'GAME_DISABLED' });
    await putBack('stargame_tictactoe');
    await page.locator('#screen-games [data-go="home"]').click();
    await openLobby(page);
    await expect(page.locator('[data-game="stargame_tictactoe"]')).toHaveAttribute('data-active', 'true');
  });

  test('English lobby with Latin digits at 390px', async ({ spawn }) => {
    const m = await spawn({ signedIn: false, mobile: true });
    await start(m, 'en', { lang: 'en' });
    await openLobby(m);
    const ttt = m.locator('[data-game="stargame_tictactoe"]');
    await expect(ttt).toContainText('Tic-tac-toe');
    await expect(ttt.locator('[data-room-card="bronze"]')).toContainText('Entry 10 Coin');
    await expect(ttt.locator('[data-room-card="silver"]')).toContainText('From level 2');
    await expect(m.locator('[data-game="stargame"]')).toContainText('30 s a round');
    await noHorizontalScroll(m);
    expect(persian(10)).toBe('۱۰');
  });
});
