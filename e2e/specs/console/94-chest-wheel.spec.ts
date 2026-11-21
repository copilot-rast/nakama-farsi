import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, open, panel, rows, savePanel, search, settlePanel } from '../../lib/console';

const k = uid();
const chestKey = 'e2e_' + k + '_chest';
const wheelKey = 'e2e_' + k + '_wheel';
const made = { users: [] as string[], prizes: [] as string[] };
let P: Player;
let coinId = '';
let small = '';
let jack = '';
let fund = '';

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

async function pickIn(scope: Locator, text: string, exact?: RegExp) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = exact ? pop.locator('[role="option"]').filter({ has: scope.page().locator('span.font-mono', { hasText: exact }) }).first() : pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

function errorOf(res: { data: any }) {
  try { return JSON.parse(res.data.message); } catch { return { code: res.data && res.data.message }; }
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-cw-' + k, ('cw' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  coinId = (await byKey('currencies', 'coin')).id;
  small = (await eco.post('/prizes', { key: 'e2e_' + k + '_small', name: { fa: 'سکهٔ کم ' + k, en: 'Small ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }] })).id;
  jack = (await eco.post('/prizes', { key: 'e2e_' + k + '_jack', name: { fa: 'جایزهٔ بزرگ ' + k, en: 'Jackpot ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 500 }] })).id;
  fund = (await eco.post('/prizes', { key: 'e2e_' + k + '_fund', name: { fa: 'شارژ ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1000 }] })).id;
  made.prizes.push(small, jack, fund);
  await eco.post('/users/' + P.id + '/grant', { prizeId: fund });
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_chance_play WHERE game_id IN (SELECT id FROM economy_chance_game WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_reward_claim WHERE kind='chance' AND owner_id IN (SELECT id FROM economy_chance_game WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (SELECT id FROM economy_prize WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/chance-games?q=e2e_' + k + '&limit=100')).items || []) expect((await eco.del('/chance-games/' + it.id)).status).toBe(200);
  for (const id of made.prizes) expect((await eco.del('/prizes/' + id)).status).toBe(200);
});

test.describe('chance games: chest slots, price ladder, pity and reveal', () => {
  test('console: a chest with an independent coin slot and a wheel with a ladder, a jackpot row, pity and reveal', async ({ page }) => {
    await open(page, '#/economy/chance-games/new');
    await settlePanel(page);
    let p = panel(page);
    await field(p, 'کلید').locator('input').fill(chestKey);
    await fillTranslated(p, 'name', { fa: 'صندوق آزمون ' + k, en: 'Test chest ' + k });
    await field(p, 'هر خانه چه می‌دهد').locator('select').selectOption('slots');
    await expect(p.locator('[data-rows="prize"]')).toBeHidden();
    await expect(p.locator('[data-ch-slots-hint]')).toBeVisible();
    const slot = p.locator('[data-rows="slots"] [data-row="0"]');
    await slot.locator('[data-col="chance"] input').fill('100');
    await slot.locator('[data-col="kind"] select').selectOption('currency');
    await pickIn(slot.locator('[data-col="currencyId"]'), 'coin', /^coin$/);
    await slot.locator('[data-col="min"] input').fill('120');
    await slot.locator('[data-col="max"] input').fill('180');
    await savePanel(page);
    const chest = await byKey('chance-games', chestKey);
    expect(chest).toMatchObject({ prizeType: 'slots', rerollEmpty: true, revealAck: false, pityThreshold: null });
    expect(chest.prize).toEqual([{ chance: 100, kind: 'currency', currencyId: coinId, min: 120, max: 180, prizeId: null, prizeGroupId: null }]);
    await search(page, chestKey);
    await expect(rows(page).first().locator('[data-ch-rows]')).toHaveText('۱ خانهٔ مستقل');

    await go(page, '#/economy/chance-games/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill(wheelKey);
    await fillTranslated(p, 'name', { fa: 'گردونهٔ پلکانی ' + k, en: 'Ladder wheel ' + k });
    const wrows = p.locator('[data-rows="prize"] [data-row]');
    await pickIn(wrows.nth(0).locator('[data-col="prizeId"]'), 'سکهٔ کم ' + k);
    await pickIn(wrows.nth(1).locator('[data-col="prizeId"]'), 'جایزهٔ بزرگ ' + k);
    await field(p, 'ردیف جایزهٔ بزرگ').locator('input').fill('2');
    await field(p, 'تضمین جایزهٔ بزرگ بعد از').locator('input').fill('40');
    await field(p, 'نوبت رایگان هر چند دقیقه').locator('input').fill('');
    await pickIn(field(p, 'ارز بلیت'), 'coin', /^coin$/);
    await field(p, 'قیمت بلیت').locator('input').fill('10');
    await field(p, 'پلکان قیمت بلیت در روز').locator('input').fill('10, 15, 20');
    await p.locator('label.x-toggle', { hasText: 'نمایش نتیجه تا تأیید بازیکن' }).click();
    await savePanel(page);
    const wheel = await byKey('chance-games', wheelKey);
    expect(wheel).toMatchObject({ prizeType: 'prize', priceLadder: [10, 15, 20], pityThreshold: 40, revealAck: true, intervalTime: null, price: { currencyId: coinId, amount: 10 } });
    expect(wheel.prize[1]).toMatchObject({ prizeId: jack, jackpot: true });
    expect(wheel.prize[0].jackpot).toBeUndefined();
    await search(page, wheelKey);
    const row = rows(page).first();
    await expect(row.locator('[data-ch-ladder]')).toContainText('۱۰ › ۱۵ › ۲۰');
    await expect(row.locator('[data-ch-pity]')).toContainText('۴۰');
    const noJack = await consoleApi('PUT', '/v2/console/economy/chance-games/' + wheel.id, Object.assign({}, wheel, { prize: wheel.prize.map((r: any) => ({ chance: r.chance, prizeId: r.prizeId || '' })) }));
    expect(noJack.status).toBe(400);
    expect(noJack.data.error).toBe('PITY_NEEDS_JACKPOT');
  });

  test('player: chest pays a coin amount in the range, ladder prices climb, expected cost, reveal waits for the ack', async () => {
    const chest = await byKey('chance-games', chestKey);
    const before = (await rpc(P, 'economy_balances', {})).data.wallet.coin;
    const opened = await rpc(P, 'economy_chance_play', { id: chest.id, lang: 'fa' });
    expect(opened.status).toBe(200);
    expect(opened.data.status).toBe('done');
    const oc = opened.data.outcome;
    expect(oc).toHaveLength(1);
    expect(oc[0]).toMatchObject({ slot: 0, kind: 'currency', currencyKey: 'coin' });
    expect(oc[0].amount).toBeGreaterThanOrEqual(120);
    expect(oc[0].amount).toBeLessThanOrEqual(180);
    expect(opened.data.wallet.coin).toBe(before + oc[0].amount);
    const hist = (await rpc(P, 'economy_chance_history', { id: chest.id })).data.items[0];
    expect(hist.outcome[0].amount).toBe(oc[0].amount);

    const wheel = await byKey('chance-games', wheelKey);
    const view = (await rpc(P, 'economy_chance_game', { id: wheel.id })).data.game;
    expect(view).toMatchObject({ revealAck: true, pendingReveal: null, price: { amount: 10 } });
    const wrong = await rpc(P, 'economy_chance_play_with_ticket', { id: wheel.id, expectedCost: 99 });
    expect(wrong.status).toBe(400);
    expect(errorOf(wrong)).toMatchObject({ code: 'CHANCE_PRICE_CHANGED', status: 409, meta: { amount: 10, expected: 99 } });
    const spent: number[] = [];
    let last = '';
    for (const cost of [10, 15, 20, 20]) {
      const r = await rpc(P, 'economy_chance_play_with_ticket', { id: wheel.id, expectedCost: cost });
      expect(r.status).toBe(200);
      expect(r.data.needsAck).toBe(true);
      spent.push(r.data.spent.amount);
      last = r.data.playId;
    }
    expect(spent).toEqual([10, 15, 20, 20]);
    const after = (await rpc(P, 'economy_chance_game', { id: wheel.id })).data.game;
    expect(after.pendingReveal.playId).toBe(last);
    expect(after.price.amount).toBe(20);
    expect(after.ticketsToday).toBe(4);
    const ack = await rpc(P, 'economy_chance_ack', { playId: last });
    expect(ack.data).toMatchObject({ acknowledged: true, playId: last });
    expect((await rpc(P, 'economy_chance_game', { id: wheel.id })).data.game.pendingReveal.playId).not.toBe(last);
    expect(errorOf(await rpc(P, 'economy_chance_ack', { playId: '00000000-0000-0000-0000-000000000001' })).code).toBe('PLAY_NOT_FOUND');
  });

  test('pity forces the jackpot on the third play and starts again', async () => {
    const g = await eco.post('/chance-games', { key: wheelKey + '_pity', name: { fa: 'تضمین ' + k }, isActive: true, prizeType: 'prize', price: { currencyId: coinId, amount: 1 }, pityThreshold: 3,
      prize: [{ chance: 99.9999, prizeId: small }, { chance: 0.0001, prizeId: jack, jackpot: true }] });
    const slots: number[] = [];
    for (let i = 0; i < 6; i++) slots.push((await rpc(P, 'economy_chance_play_with_ticket', { id: g.id })).data.slot);
    expect(slots).toEqual([0, 0, 1, 0, 0, 1]);
    expect(Number(sql(`SELECT count(*) FROM economy_chance_play WHERE game_id='${g.id}' AND forced`))).toBe(2);
    const view = (await rpc(P, 'economy_chance_game', { id: g.id })).data.game;
    expect(view.pityThreshold).toBeUndefined();
    expect(view.prize[1].jackpot).toBe(true);
  });
});
