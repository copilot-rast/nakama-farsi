import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';

const k = uid();
const made = { users: [] as string[], offers: [] as string[], entries: [] as string[], products: [] as string[], prizes: [] as string[], tags: [] as string[] };
let settings: any;
let savedRow = '';
let tag: any;
let coinPrize: any;
let P: Player;
let other: Player;

test.describe.configure({ mode: 'serial' });

async function product(key: string, coins: number) {
  const prize = await eco.post('/prizes', { key: 'e2e_' + k + '_pz_' + key, name: { fa: 'جایزه ' + key }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  const p = await eco.post('/products', { key: 'e2e_' + k + '_' + key, isActive: true, name: { fa: 'بستهٔ ' + key, en: 'Pack ' + key }, prizeId: prize.id, paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: coins } });
  made.products.push(p.id);
  return p;
}

function restoreOfferSettings(row: string) {
  if (!row) {
    sql('DELETE FROM economy_offer_settings');
    return;
  }
  const i = row.lastIndexOf('|');
  const q = (v: string) => "'" + v.replace(/'/g, "''") + "'";
  sql('UPDATE economy_offer_settings SET settings=' + q(row.slice(0, i)) + '::jsonb, updated_at=' + q(row.slice(i + 1)) + '::timestamptz');
}

function keys(list: any[]) {
  return list.map((x: any) => x.key);
}

test.beforeAll(async () => {
  savedRow = sql(`SELECT settings::text || '|' || updated_at::text FROM economy_offer_settings`);
  settings = (await eco.get('/offer-settings')).settings;
  tag = await base.post('/tags', { key: 'e2e_' + k + '_of', keyType: 'user', name: { fa: 'پیشنهاد ' + k.slice(-4) }, description: {}, isPublic: false, metadata: {}, sort: 0 });
  made.tags.push(tag.id);
  const a = await product('a', 100);
  const b = await product('b', 200);
  const c = await product('c', 300);
  const w = await product('w', 400);
  for (const [key, prod, extra] of [['a', a, { discountPercent: 50 }], ['b', b, { priceMode: 'price', coinPrice: { currencyKey: 'coin', amount: 20 }, purchaseLimit: 2 }], ['c', c, { discountPercent: 10 }]] as const) {
    const o = await eco.post('/offers', { key: 'e2e_' + k + '_' + key, name: { fa: 'پیشنهاد ' + key, en: 'Offer ' + key }, productId: prod.id, weight: 1000000, audienceTagIds: [tag.id], ...extra });
    made.offers.push(o.id);
  }
  const e = await eco.post('/entry-offers', { key: 'e2e_' + k + '_welcome', name: { fa: 'خوش‌آمد', en: 'Welcome' }, productId: w.id, discountPercent: 75, audienceTagIds: [tag.id], durationMinutes: 60, sort: -1000 });
  made.entries.push(e.id);
  coinPrize = await eco.post('/prizes', { key: 'e2e_' + k + '_coins', name: { fa: 'سکه' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1000 }], tagIds: [], metadata: {} });
  made.prizes.push(coinPrize.id);
  await eco.put('/offer-settings', { ...settings, dailyEnabled: true, entryEnabled: true, count: 2, weighting: 'priority', rerollEnabled: true, rerollLimit: 1, rerollCost: { currencyKey: 'coin', amount: 15 }, rerollCostStep: 0, entryTimerStart: 'list', entryMaxActive: 1, showSoldOut: true });
  P = await devicePlayer('e2e-of-' + k, ('of' + k).replace(/_/g, '').slice(0, 20));
  other = await devicePlayer('e2e-of2-' + k);
  made.users.push(P.id, other.id);
  await updateAccount(P, { lang_tag: 'fa' });
  await base.put('/players/' + P.id + '/tags', { tagIds: [tag.id] });
  await eco.post('/users/' + P.id + '/grant', { prizeId: coinPrize.id });
});

test.afterAll(async () => {
  restoreOfferSettings(savedRow);
  for (const id of made.users) await deleteUser(id);
  const ids = [...made.offers, ...made.entries].map((x) => `'${x}'`).join(',');
  if (ids) sql(`DELETE FROM economy_user_offer WHERE offer_id IN (${ids}); DELETE FROM economy_offer WHERE id IN (${ids})`);
  const prods = made.products.map((x) => `'${x}'`).join(',');
  if (prods) sql(`DELETE FROM economy_payment_log WHERE payment_id IN (SELECT id FROM economy_payment WHERE product_id IN (${prods})); DELETE FROM economy_payment WHERE product_id IN (${prods}); DELETE FROM economy_user_prize WHERE source_type='payment' AND prize_id IN (SELECT prize_id FROM economy_product WHERE id IN (${prods}))`);
  for (const id of made.products) await eco.del('/products/' + id);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (${made.prizes.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
});

test.describe('personal offers over RPC', () => {
  test('a tagged player gets the same personal daily set all day with a countdown; others do not', async () => {
    const first = await rpc(P, 'economy_offers', { lang: 'fa' });
    expect(first.status).toBe(200);
    const daily = first.data.daily;
    expect(daily.enabled).toBe(true);
    expect(daily.items).toHaveLength(2);
    for (const it of daily.items) expect(it.key.startsWith('e2e_' + k)).toBe(true);
    expect(daily.secondsLeft).toBeGreaterThan(0);
    expect(daily.secondsLeft).toBeLessThanOrEqual(86400);
    expect(daily.items[0]).toMatchObject({ kind: 'daily', purchaseLimit: expect.any(Number), paymentMethods: ['in_app'], soldOut: false, started: true });
    expect(daily.reroll).toMatchObject({ enabled: true, used: 0, left: 1, cost: { currencyKey: 'coin', amount: 15 } });
    const again = await rpc(P, 'economy_offers', {});
    expect(keys(again.data.daily.items)).toEqual(keys(daily.items));
    const theirs = await rpc(other, 'economy_offers', {});
    expect(keys(theirs.data.daily.items).filter((x: string) => x.startsWith('e2e_' + k))).toEqual([]);
    expect(keys(theirs.data.entry.items).filter((x: string) => x.startsWith('e2e_' + k))).toEqual([]);
    const a = daily.items.find((x: any) => x.key === 'e2e_' + k + '_a');
    if (a) expect(a).toMatchObject({ name: 'پیشنهاد a', discountPercent: 50, coinPrice: { currencyKey: 'coin', amount: 50 }, coinOriginal: 100 });
  });

  test('buying uses the offer price and limit through the shop flow; a reroll replaces the rest for coins', async () => {
    const list = (await rpc(P, 'economy_offers', {})).data.daily.items;
    const it = list[0];
    const before = (await rpc(P, 'economy_balances', {})).data.wallet.coin;
    const steal = await rpc(other, 'economy_offer_buy', { offerId: it.offerId, method: 'in_app' });
    expect(steal.data.message).toBe('OFFER_NOT_FOUND');
    const bought = await rpc(P, 'economy_offer_buy', { offerId: it.offerId, method: 'in_app' });
    expect(bought.status).toBe(200);
    expect(bought.data.status).toBe('success');
    const after = (await rpc(P, 'economy_balances', {})).data.wallet.coin;
    expect(before - after + 1).toBe(it.coinPrice.amount);
    expect(sql(`SELECT coin_amount || '|' || (user_offer_id = '${it.offerId}') FROM economy_payment WHERE id='${bought.data.paymentId}'`)).toBe(it.coinPrice.amount + '|true');
    if (it.purchaseLimit === 1) {
      const twice = await rpc(P, 'economy_purchase_start', { offerId: it.offerId, method: 'in_app' });
      expect(twice.data.message).toBe('OFFER_SOLD_OUT');
    }
    const wrong = await rpc(P, 'economy_offer_buy', { offerId: it.offerId, method: 'zarinpal' });
    expect(['METHOD_NOT_ALLOWED', 'OFFER_SOLD_OUT']).toContain(wrong.data.message);

    const rolled = await rpc(P, 'economy_offers_reroll', { lang: 'fa' });
    expect(rolled.status).toBe(200);
    expect(rolled.data.paid).toEqual({ currencyKey: 'coin', amount: 15 });
    const rk = keys(rolled.data.daily.items);
    expect(rk).toContain(it.key);
    expect(rk).not.toContain(list[1].key);
    expect((await rpc(P, 'economy_balances', {})).data.wallet.coin).toBe(after - 15);
    expect(rolled.data.daily.reroll).toMatchObject({ used: 1, left: 0 });
    const no = await rpc(P, 'economy_offers_reroll', {});
    expect(no.data.message).toBe('REROLL_LIMIT');
    const events = sql(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='Offer:Bought' AND e.user_id='${P.id}'`);
    expect(events).toBe('1');
  });

  test('the entry offer runs on the player own timer, can be closed and reset by the admin', async () => {
    const out = await rpc(P, 'economy_offers', { lang: 'fa' });
    const welcome = out.data.entry.items.find((x: any) => x.key === 'e2e_' + k + '_welcome');
    expect(welcome).toMatchObject({ kind: 'entry', name: 'خوش‌آمد', discountPercent: 75, coinPrice: { amount: 100 }, started: true, durationMinutes: 60 });
    expect(welcome.secondsLeft).toBeGreaterThan(3500);
    expect(welcome.secondsLeft).toBeLessThanOrEqual(3600);
    const seen = await rpc(P, 'economy_offer_seen', { offerId: welcome.offerId });
    expect(Math.abs(seen.data.secondsLeft - welcome.secondsLeft)).toBeLessThan(30);

    const card = await eco.get('/users/' + P.id + '/offers');
    const row = card.entry.find((x: any) => x.key === 'e2e_' + k + '_welcome');
    expect(row).toMatchObject({ state: 'active', kind: 'entry' });
    expect(card.daily.filter((x: any) => x.status === 'replaced')).toHaveLength(1);
    await eco.post('/users/' + P.id + '/offers/' + row.id + '/expire', {});
    const gone = await rpc(P, 'economy_offers', {});
    expect(keys(gone.data.entry.items)).not.toContain('e2e_' + k + '_welcome');
    const late = await rpc(P, 'economy_offer_buy', { offerId: welcome.offerId, method: 'in_app' });
    expect(late.data.message).toBe('OFFER_NOT_FOUND');
    await eco.post('/users/' + P.id + '/entry-offers/' + made.entries[0] + '/reset', {});
    const back = await rpc(P, 'economy_offers', {});
    const fresh = back.data.entry.items.find((x: any) => x.key === 'e2e_' + k + '_welcome');
    expect(fresh).toBeTruthy();
    expect(fresh.offerId).not.toBe(welcome.offerId);
  });

  test('preview and report in the console agree with what the player got', async () => {
    const pv = await eco.get('/offer-preview?userId=' + P.id);
    expect(pv.count).toBe(2);
    expect(pv.current.filter((x: any) => x.status === 'active')).toHaveLength(2);
    const reasons = Object.fromEntries(pv.draw.filter((x: any) => x.key.startsWith('e2e_' + k)).map((x: any) => [x.key, x.reason]));
    expect(Object.values(reasons).some((r) => r === 'bought_recently')).toBe(true);
    const pvOther = await eco.get('/offer-preview?userId=' + other.id);
    expect(pvOther.draw.filter((x: any) => x.key.startsWith('e2e_' + k)).every((x: any) => x.reason === 'audience')).toBe(true);
    const rep = await eco.get('/offer-report?days=7');
    const mine = rep.items.filter((x: any) => x.key.startsWith('e2e_' + k));
    expect(mine.reduce((n: number, x: any) => n + x.purchases, 0)).toBe(1);
    expect(mine.reduce((n: number, x: any) => n + x.shown, 0)).toBeGreaterThanOrEqual(4);
    const bought = mine.find((x: any) => x.purchases === 1);
    expect(bought.buyers).toBe(1);
    expect(bought.conversion).toBeGreaterThan(0);
    expect(Object.values(bought.coins)[0]).toBeGreaterThan(0);
  });
});
