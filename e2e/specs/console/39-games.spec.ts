import { expect, test } from '../../lib/fixtures';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, root, rows, savePanel, search, settlePanel } from '../../lib/console';

const k = uid();
const boardKey = 'e2e_' + k + '_board';
const ghostKey = 'e2e_' + k + '_ghost';
const made = { users: [] as string[], prizes: [] as string[] };
let P: Player;
let prizeId = '';
const prizeName = 'جایزهٔ تخته ' + k;

test.describe.configure({ mode: 'serial' });

function errorOf(res: { data: any }) {
  try { return JSON.parse(res.data.message); } catch { return { code: res.data && res.data.message }; }
}

async function byKey(key: string) {
  const res = await eco.get('/games?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-gm-' + k, ('gm' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
  prizeId = (await eco.post('/prizes', { key: 'e2e_' + k + '_win', name: { fa: prizeName, en: 'Board prize ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 8 }] })).id;
  made.prizes.push(prizeId);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_game_match WHERE game_key LIKE 'e2e\\_${k}%'`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (SELECT id FROM economy_prize WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/games?q=e2e_' + k + '&limit=100')).items || []) expect((await eco.del('/games/' + it.id)).status).toBe(200);
  for (const id of made.prizes) expect((await eco.del('/prizes/' + id)).status).toBe(200);
});

test.describe('games: console', () => {
  test('list shows the seeded games with their code, the registry warning for a game without code', async ({ page }) => {
    await open(page, '#/economy/games');
    await expect(root(page)).toContainText('منطق بازی در کد افزونه است');
    const ttt = rows(page).filter({ hasText: 'stargame_tictactoe' });
    await expect(ttt).toHaveCount(1);
    await expect(ttt).toContainText('۲ نفره');
    await expect(ttt).toContainText('هر نوبت ۱۵ ثانیه');
    await expect(ttt.locator('[data-gm-room]')).toHaveCount(3);
    await expect(ttt.locator('[data-gm-room="bronze"]')).toContainText('bronze: ۱۰ سکه');
    await expect(ttt.locator('[data-gm-room="silver"]')).toContainText('سطح ۲');
    await expect(ttt).toContainText('بازی زنده');
    const star = rows(page).filter({ hasText: 'ستاره‌چین' }).filter({ hasText: /^(?!.*tictactoe)/ });
    await expect(star.first()).toContainText('تک‌نفره');
    await expect(star.first()).toContainText('کد ثبت شده');
    await expect(root(page).locator('[data-gm-unconfigured]')).toHaveCount(0);
    await expect(root(page).locator('[data-gm-ok]')).toBeVisible();

    const ghost = await consoleApi('POST', '/v2/console/economy/games', { key: ghostKey, name: { fa: 'بازی بی‌کد ' + k }, isActive: true });
    expect(ghost.status).toBe(201);
    await page.reload();
    await expect(root(page).locator(`[data-gm-missing="${ghostKey}"]`)).toContainText('کدی روی این سرور برایش ثبت نشده');
    await search(page, ghostKey);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('کد ندارد');
    const reg = await consoleApi('GET', '/v2/console/economy/games-registry');
    expect(reg.data.missingHandler.map((x: any) => x.key)).toContain(ghostKey);
    expect(reg.data.items.find((x: any) => x.key === 'stargame_tictactoe')).toMatchObject({ realtime: true, configured: true, plugin: 'stargame' });
  });

  test('editor of a registered game: locked key, rooms rows and the form drawn from the game schema', async ({ page }) => {
    await open(page, '#/economy/games');
    await rows(page).filter({ hasText: 'stargame_tictactoe' }).click();
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('ویرایش بازی');
    await expect(field(p, 'کلید بازی').locator('input')).toHaveAttribute('readonly', '');
    await expect(p.locator('[data-gm-reg="realtime"]')).toContainText('افزونهٔ stargame');
    const rooms = p.locator('[data-rows="rooms"] [data-row]');
    await expect(rooms).toHaveCount(3);
    await expect(rooms.nth(0).locator('[data-col="key"] input')).toHaveValue('bronze');
    await expect(rooms.nth(0).locator('[data-col="feeAmount"] input')).toHaveValue('۱۰');
    await expect(rooms.nth(2).locator('[data-col="minLevel"] input')).toHaveValue('۳');
    const settings = p.locator('[data-gm-settings]');
    await expect(settings.locator('[data-setting]')).toHaveCount(3);
    await expect(settings.locator('[data-setting="firstMove"]')).toContainText('نوبت اول');
    await expect(settings.locator('[data-setting="firstMove"] select')).toHaveValue('random');
    await expect(settings.locator('[data-setting="waitSeconds"]')).toContainText('کمترین ۱۰');
    await settings.locator('[data-setting="waitSeconds"] input').fill('5');
    await p.locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(settings.locator('[data-setting="waitSeconds"]')).toContainText('یک عدد معتبر');
    const bad = await consoleApi('PUT', '/v2/console/economy/games/' + (await byKey('stargame_tictactoe')).id, Object.assign({}, await byKey('stargame_tictactoe'), { settings: { waitSeconds: 5 } }));
    expect(bad.status).toBe(400);
    expect(bad.data).toMatchObject({ error: 'INVALID_SETTING', field: 'settings.waitSeconds' });
    const unknown = await consoleApi('PUT', '/v2/console/economy/games/' + (await byKey('stargame_tictactoe')).id, Object.assign({}, await byKey('stargame_tictactoe'), { settings: { colour: 'red' } }));
    expect(unknown.data.error).toBe('UNKNOWN_SETTING');
    await page.keyboard.press('Escape');
    expect((await byKey('stargame_tictactoe')).settings).toMatchObject({ firstMove: 'random', waitSeconds: 60, forfeitOnLeave: true });
  });

  test('new game: players check, room rows with fee and prize, platforms, JSON settings for a key without code', async ({ page }) => {
    await open(page, '#/economy/games/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('تنظیم بازی تازه');
    await field(p, 'کلید بازی').locator('input').fill(boardKey);
    await expect(p.locator('[data-gm-reg="missing"]')).toBeVisible();
    await expect(p.locator('[data-gm-noschema]')).toBeVisible();
    await fillTranslated(p, 'name', { fa: 'تخته ' + k, en: 'Board ' + k });
    await fillTranslated(p, 'maintenance', { fa: 'تخته در تعمیر است', en: 'Board is under repair' });
    await field(p, 'کمترین بازیکن').locator('input').fill('3');
    await field(p, 'بیشترین بازیکن').locator('input').fill('2');
    await p.locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(field(p, 'بیشترین بازیکن')).toContainText('بیشترین بازیکن از کمترین کمتر است');
    await field(p, 'کمترین بازیکن').locator('input').fill('2');
    await field(p, 'زمان هر نوبت (ثانیه)').locator('input').fill('12');
    await p.locator('[data-rows-add="rooms"]').click();
    await p.locator('[data-rows-add="rooms"]').click();
    const rooms = p.locator('[data-rows="rooms"] [data-row]');
    await expect(rooms.nth(0).locator('[data-col="key"] input')).toHaveValue('bronze');
    await expect(rooms.nth(1).locator('[data-col="key"] input')).toHaveValue('silver');
    await rooms.nth(1).locator('[data-col="key"] input').fill('bronze');
    await p.locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(rooms.nth(1)).toContainText('کلید هر اتاق باید یکتا باشد');
    await rooms.nth(1).locator('[data-col="key"] input').fill('silver');
    const r0 = rooms.nth(0);
    await r0.locator('[data-col="feeCurrency"] button[aria-haspopup="listbox"]').click();
    await r0.locator('[data-col="feeCurrency"] .eco-pop input').fill('coin');
    await r0.locator('[data-col="feeCurrency"] .eco-pop [role="option"]').filter({ has: page.locator('span.font-mono', { hasText: /^coin$/ }) }).first().click();
    await r0.locator('[data-col="feeAmount"] input').fill('۵');
    await r0.locator('[data-col="prizeId"] button[aria-haspopup="listbox"]').click();
    await r0.locator('[data-col="prizeId"] .eco-pop input').fill(prizeName);
    await r0.locator('[data-col="prizeId"] .eco-pop [role="option"]', { hasText: prizeName }).first().click();
    await rooms.nth(1).locator('[data-col="minLevel"] input').fill('4');
    await p.locator('[data-gm-settings] textarea').fill('{"pieces": 15}');
    await p.locator('label', { hasText: 'وب' }).first().click();
    await savePanel(page);
    const g = await byKey(boardKey);
    expect(g).toMatchObject({ minPlayers: 2, maxPlayers: 2, turnSeconds: 12, platforms: ['web'], settings: { pieces: 15 }, isActive: true });
    expect(g.rooms).toEqual([
      { key: 'bronze', entryFee: { currencyKey: 'coin', amount: 5 }, prizeId },
      { key: 'silver', minLevel: 4 },
    ]);
    await search(page, boardKey);
    const row = rows(page).first();
    await expect(row).toContainText('bronze: ۵ سکه');
    await expect(row).toContainText('هر نوبت ۱۲ ثانیه');
    const fee = await consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, { rooms: [{ key: 'bronze', entryFee: { currencyKey: 'coin' } }] }));
    expect(fee.data).toMatchObject({ error: 'INVALID_ENTRY_FEE', field: 'rooms[0]' });
    const once = (await eco.post('/prizes', { key: 'e2e_' + k + '_once', name: { fa: 'یک‌باره ' + k }, isRepeatable: false, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }] })).id;
    made.prizes.push(once);
    const rep = await consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, { rooms: [{ key: 'bronze', prizeId: once }] }));
    expect(rep.data.error).toBe('PRIZE_NOT_REPEATABLE');
    const board = await consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, { leaderboardId: 'e2e_no_board' }));
    expect(board.data.error).toBe('LEADERBOARD_NOT_FOUND');
    const audit = sql(`SELECT string_agg(DISTINCT action, ',' ORDER BY action) FROM console_audit_log WHERE message LIKE 'economy % games%'`);
    expect(audit.split(',')).toEqual(expect.arrayContaining(['CREATE']));
  });

  test('client RPCs: lobby list per platform, join ticket, level and balance checks, maintenance when off', async () => {
    const g = await byKey(boardKey);
    const web = await rpc(P, 'economy_games', { lang: 'fa', platform: 'web', version: '1.4.0' });
    const mine = web.data.items.find((x: any) => x.key === boardKey);
    expect(mine).toMatchObject({ name: 'تخته ' + k, isActive: true, multiplayer: true, realtime: false, turnSeconds: 12 });
    expect(mine.rooms[0]).toMatchObject({ key: 'bronze', locked: false, entryFee: { currencyKey: 'coin', amount: 5, currencyName: 'سکه' }, prize: { type: 'prize', name: prizeName } });
    expect(mine.rooms[1]).toMatchObject({ key: 'silver', locked: true, minLevel: 4 });
    const android = await rpc(P, 'economy_games', { platform: 'android' });
    expect(android.data.items.find((x: any) => x.key === boardKey)).toBeUndefined();
    expect(android.data.items.find((x: any) => x.key === 'stargame_tictactoe')).toBeTruthy();
    const ticket = await rpc(P, 'economy_game_join', { key: boardKey, room: 'bronze', platform: 'web' });
    expect(ticket.status).toBe(200);
    expect(ticket.data).toMatchObject({ query: `+properties.game:${boardKey} +properties.room:bronze`, minCount: 2, maxCount: 2, stringProperties: { game: boardKey, room: 'bronze' } });
    const low = await rpc(P, 'economy_game_join', { key: boardKey, room: 'silver' });
    expect(errorOf(low)).toMatchObject({ code: 'LEVEL_TOO_LOW', meta: { minLevel: 4 } });
    const none = await rpc(P, 'economy_game_join', { key: boardKey });
    expect(errorOf(none).code).toBe('ROOM_REQUIRED');
    const ios = await rpc(P, 'economy_game_join', { key: boardKey, room: 'bronze', platform: 'ios' });
    expect(errorOf(ios).code).toBe('PLATFORM_NOT_SUPPORTED');
    const gold = await rpc(P, 'economy_game_join', { key: 'stargame_tictactoe', room: 'gold' });
    expect(errorOf(gold).code).toBe('LEVEL_TOO_LOW');
    const solo = await rpc(P, 'economy_game_join', { key: 'stargame' });
    expect(errorOf(solo).code).toBe('SINGLE_PLAYER_GAME');

    expect((await consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, { isActive: false }))).status).toBe(200);
    const off = await rpc(P, 'economy_games', { lang: 'en', platform: 'web' });
    expect(off.data.items.find((x: any) => x.key === boardKey)).toMatchObject({ isActive: false, maintenance: 'Board is under repair' });
    const shut = await rpc(P, 'economy_game_join', { key: boardKey, room: 'bronze', lang: 'fa' });
    expect(errorOf(shut)).toMatchObject({ code: 'GAME_DISABLED', meta: { message: 'تخته در تعمیر است' } });
    expect((await consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, { isActive: false, maintenance: {} }))).status).toBe(200);
    const hidden = await rpc(P, 'economy_games', { platform: 'web' });
    expect(hidden.data.items.find((x: any) => x.key === boardKey)).toBeUndefined();
    const one = await rpc(P, 'economy_game', { key: 'stargame', lang: 'en' });
    expect(one.data).toMatchObject({ key: 'stargame', name: 'Star Catcher', multiplayer: false, settings: { roundSeconds: 30, maxCoins: 30 } });
    const matches = await consoleApi('GET', '/v2/console/economy/game-matches?gameKey=' + boardKey);
    expect(matches.status).toBe(200);
  });

  test('sidebar entry, page search, English and 390px dark cards without sideways scroll', async ({ page, spawn }) => {
    await open(page, '#/economy/overview');
    await expect(page.locator('.pl-side a.pl-nav[href="#/economy/games"] svg')).toHaveCount(1);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('بازی‌ها');
    const hit = page.locator('.pl-pal-row').filter({ has: page.locator('.pl-pal-label', { hasText: /^بازی‌ها$/ }) }).first();
    await expect(hit).toBeVisible();
    await hit.click();
    await expect(page).toHaveURL(/#\/economy\/games/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/economy/games?q=stargame_tictactoe');
    const en = rows(page).filter({ hasText: 'Tic-tac-toe' });
    await expect(en).toHaveCount(1);
    await expect(en).toContainText('2 players');
    await expect(en).toContainText('Live match');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
    const m = await spawn({ signedIn: true, mobile: true, mode: 'dark' });
    await open(m, '#/economy/games');
    await expect(rows(m).filter({ hasText: 'دوز' })).toHaveCount(1);
    await noHorizontalScroll(m);
    await rows(m).filter({ hasText: 'دوز' }).click();
    await settlePanel(m);
    await expect(panel(m).locator('[data-rows="rooms"] [data-row]')).toHaveCount(3);
    await noHorizontalScroll(m);
  });
});
