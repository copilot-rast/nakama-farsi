import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, root, rows, savePanel, search, settlePanel } from '../../lib/console';

const k = uid();
const gameKey = 'e2e_' + k + '_wheel';
const groupKey = 'e2e_' + k + '_pool';
const made = { users: [] as string[], prizes: [] as string[] };
let P: Player;
const prizeA = 'سکهٔ شانس الف ' + k;
const prizeB = 'سکهٔ شانس ب ' + k;
let idA = '';
let idB = '';
let coinId = '';

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

async function pickKey(scope: Locator, key: string) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(key);
  const opt = pop.locator('[role="option"]').filter({ has: scope.page().locator('span.font-mono', { hasText: new RegExp('^' + key + '$') }) }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

function errorOf(res: { data: any }) {
  try { return JSON.parse(res.data.message); } catch { return { code: res.data && res.data.message }; }
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-ch-' + k, ('ch' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
  idA = (await eco.post('/prizes', { key: 'e2e_' + k + '_a', name: { fa: prizeA, en: 'Chance coins A ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 9 }] })).id;
  idB = (await eco.post('/prizes', { key: 'e2e_' + k + '_b', name: { fa: prizeB, en: 'Chance coins B ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 4 }] })).id;
  made.prizes.push(idA, idB);
  coinId = (await byKey('currencies', 'coin')).id;
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_chance_play WHERE game_id IN (SELECT id FROM economy_chance_game WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_reward_claim WHERE kind='chance' AND owner_id IN (SELECT id FROM economy_chance_game WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (SELECT id FROM economy_prize WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/chance-games?q=e2e_' + k + '&limit=100')).items || []) expect((await eco.del('/chance-games/' + it.id)).status).toBe(200);
  for (const it of (await eco.get('/prize-groups?q=e2e_' + k + '&limit=100')).items || []) expect((await eco.del('/prize-groups/' + it.id)).status).toBe(200);
  for (const id of made.prizes) expect((await eco.del('/prizes/' + id)).status).toBe(200);
});

test.describe('chance games: console', () => {
  test('prize group editor: weights with live shares, duplicate refused, list row', async ({ page }) => {
    await open(page, '#/economy/prize-groups');
    await expect(root(page)).toContainText('فقط یکی از جایزه‌هایش');
    await expect(rows(page).filter({ hasText: 'star_box_coins' })).toHaveCount(1);
    await go(page, '#/economy/prize-groups/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('گروه جایزهٔ تازه');
    await field(p, 'کلید').locator('input').fill(groupKey);
    await fillTranslated(p, 'name', { fa: 'سبد ' + k, en: 'Pool ' + k });
    const members = p.locator('[data-rows="members"] [data-row]');
    await pickIn(members.nth(0).locator('[data-col="prizeId"]'), prizeA);
    await members.nth(0).locator('[data-col="weight"] input').fill('3');
    await p.locator('[data-rows-add="members"]').click();
    await pickIn(members.nth(1).locator('[data-col="prizeId"]'), prizeA);
    await savePanel(page).catch(() => null);
    await expect(members.nth(1)).toContainText('فقط یک بار در گروه');
    await pickIn(members.nth(1).locator('[data-col="prizeId"]'), prizeB);
    const shares = p.locator('[data-pg-shares]');
    await expect(shares.locator('[data-pg-share="0"]')).toContainText('۷۵٪');
    await expect(shares.locator('[data-pg-share="1"]')).toContainText('۲۵٪');
    await savePanel(page);
    const g = await byKey('prize-groups', groupKey);
    expect(g.members.map((m: any) => [m.prizeId, m.weight])).toEqual([[idA, 3], [idB, 1]]);
    await search(page, groupKey);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('۲ جایزه');
    await expect(rows(page).first().locator('[data-pg-games]')).toHaveText('۰ بازی شانس');
    const empty = await consoleApi('PUT', '/v2/console/economy/prize-groups/' + g.id, Object.assign({}, g, { members: [] }));
    expect(empty.status).toBe(400);
    expect(empty.data.error).toBe('PRIZE_GROUP_EMPTY');
  });

  test('chance game editor: chances table with a live sum of 100, prize or group slots, free turn and ticket price', async ({ page }) => {
    await open(page, '#/economy/chance-games');
    await expect(root(page)).toContainText('گردونه، جعبهٔ شانس');
    await expect(rows(page).filter({ hasText: 'star_wheel_daily' })).toHaveCount(1);
    await go(page, '#/economy/chance-games/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('بازی شانس تازه');
    await field(p, 'کلید').locator('input').fill(gameKey);
    await fillTranslated(p, 'name', { fa: 'گردونهٔ آزمون ' + k, en: 'Test wheel ' + k });
    const slots = p.locator('[data-rows="prize"] [data-row]');
    await expect(slots).toHaveCount(3);
    const sum = p.locator('[data-ch-sum]');
    await expect(sum).toHaveAttribute('data-ch-sum', 'ok');
    await expect(sum).toContainText('جمع شانس‌ها: ۱۰۰ از ۱۰۰');
    await expect(slots.nth(0).locator('[data-col="prizeGroupId"]')).toBeHidden();
    await pickIn(slots.nth(0).locator('[data-col="prizeId"]'), prizeA);
    await pickIn(slots.nth(1).locator('[data-col="prizeId"]'), prizeB);
    await slots.nth(0).locator('[data-col="chance"] input').fill('20');
    await expect(sum).toHaveAttribute('data-ch-sum', 'bad');
    await expect(sum).toContainText('۹۰ از ۱۰۰');
    await expect(sum).toContainText('۱۰ مانده');
    await savePanel(page).catch(() => null);
    await expect(slots.nth(2)).toContainText('جمع شانس‌ها باید دقیقاً ۱۰۰ باشد');
    await slots.nth(0).locator('[data-col="chance"] input').fill('25.5');
    await slots.nth(2).locator('[data-col="chance"] input').fill('34.5');
    await expect(sum).toHaveAttribute('data-ch-sum', 'ok');
    await expect(sum.locator('[data-ch-bar] > span')).toHaveCount(3);
    await field(p, 'نوبت رایگان هر چند دقیقه').locator('input').fill('60');
    await pickKey(field(p, 'ارز بلیت'), 'coin');
    await field(p, 'قیمت بلیت').locator('input').fill('5');
    await savePanel(page);
    const g = await byKey('chance-games', gameKey);
    expect(g).toMatchObject({ prizeType: 'prize', intervalTime: 60, isActive: true, price: { currencyId: coinId, amount: 5 } });
    expect(g.prize).toEqual([{ chance: 25.5, prizeId: idA, prizeGroupId: null }, { chance: 40, prizeId: idB, prizeGroupId: null }, { chance: 34.5, prizeId: null, prizeGroupId: null }]);

    await search(page, gameKey);
    await expect(rows(page)).toHaveCount(1);
    const row = rows(page).first();
    await expect(row).toContainText('گردونهٔ آزمون ' + k);
    await expect(row).toContainText('رایگان هر ۱ ساعت');
    await expect(row.locator('[data-ch-price]')).toContainText('بلیت: ۵');
    await expect(row.locator('[data-ch-rows]')).toHaveText('۳ خانه، ۲ با جایزه');
    await row.locator('td').first().click();
    await settlePanel(page);
    const q = panel(page);
    await expect(q).toContainText('ویرایش بازی شانس');
    await expect(field(q, 'کلید').locator('input')).toHaveJSProperty('readOnly', true);
    await field(q, 'هر خانه چه می‌دهد').locator('select').selectOption('prizeGroup');
    const qs = q.locator('[data-rows="prize"] [data-row]');
    await expect(qs.nth(0).locator('[data-col="prizeId"]')).toBeHidden();
    await expect(qs.nth(0).locator('[data-col="prizeGroupId"]')).toBeVisible();
    await pickIn(qs.nth(0).locator('[data-col="prizeGroupId"]'), 'سبد ' + k);
    await field(q, 'نوبت رایگان هر چند دقیقه').locator('input').fill('');
    await savePanel(page);
    const h = await byKey('chance-games', gameKey);
    const group = await byKey('prize-groups', groupKey);
    expect(h.prizeType).toBe('prizeGroup');
    expect(h.intervalTime).toBeNull();
    expect(h.prize).toEqual([{ chance: 25.5, prizeId: null, prizeGroupId: group.id }, { chance: 40, prizeId: null, prizeGroupId: null }, { chance: 34.5, prizeId: null, prizeGroupId: null }]);
    await expect(rows(page).first().locator('[data-ch-rows]')).toHaveText('۳ خانه، ۱ با جایزه');

    const bad = await consoleApi('PUT', '/v2/console/economy/chance-games/' + h.id, Object.assign({}, h, { prize: [{ chance: 30, prizeGroupId: group.id }, { chance: 60 }] }));
    expect(bad.status).toBe(400);
    expect(bad.data).toMatchObject({ error: 'TOTAL_CHANCE_MUST_BE_100', field: 'prize', meta: { totalChance: 90 } });
    const neither = await consoleApi('PUT', '/v2/console/economy/chance-games/' + h.id, Object.assign({}, h, { intervalTime: null, price: null }));
    expect(neither.data.error).toBe('INTERVAL_OR_PRICE_REQUIRED');
    const used = await consoleApi('DELETE', '/v2/console/economy/prize-groups/' + group.id);
    expect(used.status).toBe(409);
    expect(used.data.error).toBe('IN_USE');
    const audit = sql(`SELECT string_agg(DISTINCT action, ',' ORDER BY action) FROM console_audit_log WHERE message LIKE 'economy % chance-games%'`);
    expect(audit.split(',')).toEqual(expect.arrayContaining(['CREATE', 'UPDATE']));
  });

  test('player card: plays from the game with wins, nothing and tickets, the RPC error shape', async ({ page }) => {
    const free = await eco.post('/chance-games', { key: gameKey + '_free', name: { fa: 'شانس رایگان ' + k, en: 'Free luck ' + k }, isActive: true, prizeType: 'prize', intervalTime: 60, prize: [{ chance: 100, prizeId: idA }] });
    const blank = await eco.post('/chance-games', { key: gameKey + '_blank', name: { fa: 'شانس پوچ ' + k, en: 'Blank luck ' + k }, isActive: true, prizeType: 'prize', price: { currencyId: coinId, amount: 3 }, prize: [{ chance: 100 }] });
    const before = (await rpc(P, 'economy_balances', {})).data.wallet.coin;
    const won = await rpc(P, 'economy_chance_play', { id: free.id, lang: 'fa' });
    expect(won.status).toBe(200);
    expect(won.data).toMatchObject({ code: 'SUCCESS', prizeId: idA, mode: 'free', slot: 0 });
    const again = await rpc(P, 'economy_chance_play', { id: free.id });
    expect(again.status).toBe(403);
    expect(again.data.code).toBe(7);
    const wait = errorOf(again);
    expect(wait).toMatchObject({ code: 'WAITING_FOR_TURN', status: 403 });
    expect(wait.meta.nextPlayTime).toBeGreaterThan(3500);
    expect((await rpc(P, 'economy_chance_can_play', { id: free.id })).data).toMatchObject({ code: 'SUCCESS', canPlay: false });
    expect(errorOf(await rpc(P, 'economy_chance_play_with_ticket', { id: free.id })).code).toBe('TICKET_PLAY_IS_NOT_SUPPORT');
    expect(errorOf(await rpc(P, 'economy_chance_next_play_time', { id: blank.id })).code).toBe('INTERVAL_TIME_IS_NULL');
    const lost = await rpc(P, 'economy_chance_play_with_ticket', { id: blank.id });
    expect(lost.data).toMatchObject({ code: 'SUCCESS', prizeId: null, spent: { currencyKey: 'coin', amount: 3 } });
    expect((await rpc(P, 'economy_balances', {})).data.wallet.coin).toBe(before + 9 - 3);
    const events = sql(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='ChanceGame:Played' AND e.user_id='${P.id}'`);
    expect(Number(events)).toBe(2);

    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-chance]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('h2')).toHaveText('بازی‌های شانس بازیکن');
    await expect(card.locator('[data-pch-counts]')).toContainText('۲ بار بازی');
    const lines = card.locator('[data-pch]');
    await expect(lines).toHaveCount(2);
    await expect(card.locator('[data-won="true"]')).toContainText('برد: ' + prizeA);
    await expect(card.locator('[data-won="true"]')).toContainText('رایگان');
    await expect(card.locator('[data-won="false"]')).toContainText('پوچ');
    await expect(card.locator('[data-won="false"] [data-pch-spent]')).toContainText('۳');
    const hist = await consoleApi('GET', '/v2/console/economy/chance-plays?gameId=' + free.id);
    expect(hist.data.total).toBe(1);
    expect(hist.data.items[0]).toMatchObject({ userId: P.id, mode: 'free', prizeId: idA, status: 'done' });
    const usedPrize = await consoleApi('DELETE', '/v2/console/economy/prizes/' + idA);
    expect(usedPrize.status).toBe(409);
  });

  test('page search, English and 390px cards without sideways scroll', async ({ page, spawn }) => {
    await open(page, '#/economy/overview');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('بازی شانس');
    await expect(page.locator('.pl-pal-row').first()).toContainText('بازی شانس');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/chance-games/);
    await expect(page.locator('.pl-pal-input')).toHaveCount(0);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('گروه‌های جایزه');
    await expect(page.locator('.pl-pal-row').first()).toContainText('گروه‌های جایزه');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/prizes\/groups/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/economy/chance-games?q=' + gameKey);
    const en = rows(page).filter({ hasText: 'Test wheel ' + k });
    await expect(en).toHaveCount(1);
    await expect(en.locator('[data-ch-price]')).toContainText('Ticket: 5');
    await expect(en.locator('[data-ch-rows]')).toHaveText('3 slots, 1 with a prize');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
    const m = await spawn({ signedIn: true, mobile: true, mode: 'dark' });
    await open(m, '#/economy/chance-games');
    await expect(rows(m).filter({ hasText: 'گردونهٔ روزانه' })).toHaveCount(1);
    await noHorizontalScroll(m);
    await go(m, '#/economy/prize-groups');
    await noHorizontalScroll(m);
  });
});
