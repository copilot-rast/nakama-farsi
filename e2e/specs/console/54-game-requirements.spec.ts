import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, go, noHorizontalScroll, open, panel, pickRef, savePanel, search, rows, settlePanel } from '../../lib/console';

const k = uid();
const gameKey = 'e2e_' + k + '_arena';
const made = { users: [] as string[], prizes: [] as string[], products: [] as string[], items: [] as string[], metas: [] as string[], tags: [] as string[] };
let P: Player;
let passId = '';
let ticketMeta = '';
let vipTag = '';
let cheatTag = '';
let productId = '';
const passName = 'مجوز آرنا ' + k.slice(-4);
const vipName = 'ویژهٔ آرنا ' + k.slice(-4);

test.describe.configure({ mode: 'serial' });

function errorOf(res: { data: any }) {
  try { return JSON.parse(res.data.message); } catch { return { code: res.data && res.data.message }; }
}

async function game() {
  const res = await eco.get('/games?q=' + encodeURIComponent(gameKey) + '&limit=20');
  return (res.items || []).find((x: any) => x.key === gameKey);
}

function lobbyRoom(list: any, key: string) {
  const g = list.items.find((x: any) => x.key === gameKey);
  return g ? (g.rooms || []).find((r: any) => r.key === key) : undefined;
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-req-' + k, ('rq' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
  const anyItem = (await eco.get('/items?limit=1')).items[0];
  passId = (await eco.post('/items', { key: 'e2e_' + k + '_pass', typeId: anyItem.typeId, name: { fa: passName, en: 'Arena pass ' + k.slice(-4) }, rarity: 'common', isActive: true })).id;
  made.items.push(passId);
  ticketMeta = (await eco.post('/metas', { key: 'e2e_' + k + '_tk', name: { fa: 'بلیت آرنا', en: 'Arena ticket' }, defaultValue: 0 })).id;
  made.metas.push(ticketMeta);
  vipTag = (await base.post('/tags', { key: 'e2e_' + k + '_vip', keyType: 'user', name: { fa: vipName, en: 'Arena VIP' }, description: {}, isPublic: true, metadata: {}, sort: 0 })).id;
  cheatTag = (await base.post('/tags', { key: 'e2e_' + k + '_bad', keyType: 'user', name: { fa: 'متقلب ' + k.slice(-4) }, description: {}, isPublic: false, metadata: {}, sort: 0 })).id;
  made.tags.push(vipTag, cheatTag);
  const prize = (await eco.post('/prizes', { key: 'e2e_' + k + '_pz', name: { fa: 'مجوز' }, isRepeatable: true, items: [{ type: 'item', key: 'e2e_' + k + '_pass', action: 'give', value: 1 }] })).id;
  made.prizes.push(prize);
  productId = (await eco.post('/products', { key: 'e2e_' + k + '_buy', isActive: true, name: { fa: 'خرید مجوز ' + k.slice(-4), en: 'Buy pass' }, prizeId: prize, paymentMethods: ['zibal'], priceToman: 40000 })).id;
  made.products.push(productId);
  const g = await consoleApi('POST', '/v2/console/economy/games', { key: gameKey, isActive: true, name: { fa: 'آرنا ' + k, en: 'Arena ' + k }, minPlayers: 2, maxPlayers: 2,
    rooms: [{ key: 'open', name: { fa: 'آزاد' } }, { key: 'arena', name: { fa: 'آرنا' } }, { key: 'club', name: { fa: 'باشگاه' } }] });
  expect(g.status).toBe(201);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_game_match WHERE game_key LIKE 'e2e\\_${k}%'`);
  for (const it of (await eco.get('/games?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/games/' + it.id)).status).toBe(200);
  for (const id of made.products) expect((await eco.del('/products/' + id)).status).toBe(200);
  for (const id of made.prizes) expect((await eco.del('/prizes/' + id)).status).toBe(200);
  sql(`DELETE FROM economy_user_item WHERE item_id IN (SELECT id FROM economy_item WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_user_meta WHERE meta_id IN (SELECT id FROM economy_meta WHERE key LIKE 'e2e\\_${k}%')`);
  for (const id of made.items) expect((await eco.del('/items/' + id)).status).toBe(200);
  for (const id of made.metas) expect((await eco.del('/metas/' + id)).status).toBe(200);
  for (const id of made.tags) expect((await base.del('/tags/' + id)).status).toBe(200);
});

test.describe('game room entry requirements', () => {
  test('room editor: folded requirements with a summary, item ticket, meta ticket, paths and a refusal message; game-level excluded tag', async ({ page }) => {
    await open(page, '#/economy/games?q=' + gameKey);
    await rows(page).filter({ hasText: gameKey }).click();
    await settlePanel(page);
    const p = panel(page);
    const roomsEl = p.locator('[data-rows="rooms"] [data-row]');
    await expect(roomsEl).toHaveCount(3);
    const arena = roomsEl.nth(1);
    const req = arena.locator('[data-gm-req="room"]');
    await expect(req.locator('[data-gm-req-summary]')).toHaveText('بدون شرط جداگانه');
    await expect(req.locator('[data-gm-req-panel]')).toBeHidden();
    await req.locator('[data-gm-req-toggle]').click();
    await expect(req.locator('[data-gm-req-panel]')).toBeVisible();
    await pickRef(req, 'آیتم‌های لازم', 'e2e_' + k + '_pass', passName);
    await expect(req.locator('[data-gm-req-summary]')).toHaveText('۱ آیتم (یکی کافی است)');
    await req.locator('label.x-toggle', { hasText: 'آیتم با ورود مصرف شود (بلیت)' }).click();
    await expect(req.locator('[data-gm-req-summary]')).toContainText('بلیت مصرف می‌شود');
    await expect(field(req, 'بلیت وقتی بازیکن پیش از شروع برود یا بازی لغو شود')).toBeVisible();
    await req.locator('label', { hasText: 'بازی با ربات' }).first().click();
    await expect(req.locator('[data-gm-req-summary]')).toContainText('حریف‌یابی، اتاق دوستانه');
    await req.locator('[data-gm-req-toggle]').click();
    await expect(req.locator('[data-gm-req-panel]')).toBeHidden();

    const club = roomsEl.nth(2).locator('[data-gm-req="room"]');
    await club.locator('[data-gm-req-toggle]').click();
    await pickRef(club, 'برچسب‌های لازم بازیکن', 'e2e_' + k + '_vip', vipName);
    await club.locator('[data-rows-add="metas"]').click();
    const meta = club.locator('[data-rows="metas"] [data-row]').first();
    await meta.locator('[data-col="metaId"] button[aria-haspopup="listbox"]').click();
    await meta.locator('[data-col="metaId"] .eco-pop input').fill('e2e_' + k + '_tk');
    await meta.locator('[data-col="metaId"] .eco-pop [role="option"]').first().click();
    await meta.locator('[data-col="consume"] input').fill('1');
    await field(club, 'اتاق برای بازیکنی که شرط را ندارد').locator('select').selectOption('hide');
    await expect(club.locator('[data-gm-req-summary]')).toHaveText(/۱ برچسب \(یکی کافی است\) · ۱ شرط متا · بلیت مصرف می‌شود · برای بی‌شرط‌ها پنهان/);

    const whole = p.locator('[data-gm-req="game"]');
    await expect(whole.locator('[data-gm-req-summary]')).toHaveText('بدون شرط؛ همه می‌توانند وارد شوند');
    await whole.locator('[data-gm-req-toggle]').click();
    await pickRef(whole, 'برچسب‌های ممنوع', 'e2e_' + k + '_bad');
    await pickRef(whole, 'برچسب‌های لازم بازیکن', 'e2e_' + k + '_bad');
    await p.locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(whole).toContainText('یک برچسب نمی‌تواند هم لازم باشد هم ممنوع');
    await whole.locator('.eco-chip-x').first().click();
    await savePanel(page);

    const g = await game();
    expect(g.requirements).toMatchObject({ excludeTagIds: [cheatTag], lock: 'show' });
    expect(g.requirements.tagIds || []).toEqual([]);
    expect(g.rooms[0].requirements).toBeUndefined();
    expect(g.rooms[1].requirements).toMatchObject({ itemIds: [passId], itemsMode: 'any', consumeItems: true, paths: ['matchmaking', 'friend'], refund: 'return' });
    expect(g.rooms[2].requirements).toMatchObject({ tagIds: [vipTag], metas: [{ metaId: ticketMeta, operator: 'gte', value: 1, consume: 1 }], lock: 'hide' });
    await search(page, gameKey);
    await expect(rows(page).first().locator('[data-gm-room-req]')).toHaveCount(2);
  });

  test('server checks: a starter ticket, an unknown path and a tag both ways are refused with the field', async () => {
    const g = await game();
    const starter = (await eco.get('/items?isStarter=true&limit=1')).items[0];
    const put = (rooms: any[], requirements?: any) => consoleApi('PUT', '/v2/console/economy/games/' + g.id, Object.assign({}, g, { rooms, requirements: requirements || g.requirements }));
    if (starter) {
      const s = await put([{ key: 'open', requirements: { itemIds: [starter.id], consumeItems: true } }]);
      expect(s.data).toMatchObject({ error: 'STARTER_TICKET', field: 'rooms[0].requirements.consumeItems' });
    }
    const path = await put([{ key: 'open', requirements: { itemIds: [passId], paths: ['arena'] } }]);
    expect(path.data).toMatchObject({ error: 'INVALID_PATH', field: 'rooms[0].requirements.paths' });
    const both = await put(g.rooms, { tagIds: [vipTag], excludeTagIds: [vipTag] });
    expect(both.data).toMatchObject({ error: 'TAG_BOTH_WAYS', field: 'requirements.excludeTagIds' });
    const inUse = await eco.del('/items/' + passId);
    expect(inUse.status).toBe(409);
    expect(inUse.data.error).toBe('IN_USE');
  });

  test('economy_games shows requirements, eligibility per path and products; economy_game_join refuses with what is missing', async () => {
    let list = (await rpc(P, 'economy_games', { lang: 'fa' })).data;
    expect(lobbyRoom(list, 'club')).toBeUndefined();
    const arena = lobbyRoom(list, 'arena');
    expect(arena).toMatchObject({ eligible: false, locked: true, levelLocked: false, consumes: true, eligibleBy: { matchmaking: false, bot: true, friend: false } });
    expect(arena.requirements.items).toMatchObject({ mode: 'any', consume: true, list: [{ key: 'e2e_' + k + '_pass', name: passName, owned: false }] });
    expect(arena.missing).toEqual([{ type: 'item', mode: 'any', items: [{ key: 'e2e_' + k + '_pass', name: passName, iconUrl: '' }] }]);
    expect(arena.products.map((x: any) => x.key)).toEqual(['e2e_' + k + '_buy']);
    expect(arena.products[0]).toMatchObject({ priceToman: 40000, satisfies: ['item:e2e_' + k + '_pass'] });
    expect(lobbyRoom(list, 'open')).toMatchObject({ eligible: true, locked: false, products: [] });
    const bot = (await rpc(P, 'economy_games', { path: 'bot' })).data;
    expect(lobbyRoom(bot, 'arena')).toMatchObject({ eligible: true, locked: false });
    expect(errorOf(await rpc(P, 'economy_games', { path: 'arena' })).code).toBe('INVALID_PATH');

    const refused = await rpc(P, 'economy_game_join', { key: gameKey, room: 'arena', lang: 'fa' });
    const e = errorOf(refused);
    expect(e).toMatchObject({ code: 'REQUIREMENTS_NOT_MET', status: 403, meta: { game: gameKey, room: 'arena', path: 'matchmaking', message: 'برای ورود به این اتاق شرط‌هایش را نداری.' } });
    expect(e.meta.products[0].id).toBe(productId);
    expect((await rpc(P, 'economy_game_join', { key: gameKey, room: 'arena', path: 'bot' })).status).toBe(200);

    sql(`INSERT INTO economy_user_item (user_id, item_id, source) VALUES ('${P.id}', '${passId}', 'admin')`);
    const ok = await rpc(P, 'economy_game_join', { key: gameKey, room: 'arena' });
    expect(ok.status).toBe(200);
    expect(ok.data).toMatchObject({ path: 'matchmaking', consumes: true });

    await base.put('/players/' + P.id + '/tags', { tagIds: [vipTag] });
    list = (await rpc(P, 'economy_games', { lang: 'en' })).data;
    expect(lobbyRoom(list, 'club')).toBeUndefined();
    const noTicket = errorOf(await rpc(P, 'economy_game_join', { key: gameKey, room: 'club', lang: 'en' }));
    expect(noTicket).toMatchObject({ code: 'REQUIREMENTS_NOT_MET', meta: { missing: [{ type: 'meta', key: 'e2e_' + k + '_tk', name: 'Arena ticket', operator: 'gte', value: 1, have: 0, consume: 1 }] } });
    sql(`INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ('${P.id}', '${ticketMeta}', 2)`);
    list = (await rpc(P, 'economy_games', { lang: 'en' })).data;
    expect(lobbyRoom(list, 'club')).toMatchObject({ eligible: true, locked: false });

    await base.put('/players/' + P.id + '/tags', { tagIds: [vipTag, cheatTag] });
    const banned = errorOf(await rpc(P, 'economy_game_join', { key: gameKey, room: 'open' }));
    expect(banned.code).toBe('REQUIREMENTS_NOT_MET');
    expect(banned.meta.missing).toEqual([{ type: 'excludedTag', tags: [] }]);
    await base.put('/players/' + P.id + '/tags', { tagIds: [] });
  });

  test('English summary and the requirements panel at 390px without sideways scroll', async ({ page }) => {
    await open(page, '#/economy/games?q=' + gameKey);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await page.setViewportSize({ width: 390, height: 844 });
    await go(page, '#/economy/games?q=' + gameKey);
    await rows(page).filter({ hasText: gameKey }).click();
    await settlePanel(page);
    const p = panel(page);
    const arena = p.locator('[data-rows="rooms"] [data-row]').nth(1).locator('[data-gm-req="room"]');
    await expect(arena.locator('[data-gm-req-summary]')).toContainText('Items: 1 (any one) · uses a ticket');
    await arena.locator('[data-gm-req-toggle]').click();
    await expect(arena.locator('[data-gm-req-panel]')).toBeVisible();
    await expect(arena).toContainText('Use the item up on entry (ticket)');
    await noHorizontalScroll(page);
    await page.keyboard.press('Escape');
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
  });
});
