import { expect, test } from '../../lib/fixtures';
import { API_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';

const k = uid();
const made = { users: [] as string[], products: [] as string[], prizes: [] as string[], items: [] as string[], types: [] as string[] };
let savedRow = '';
let A: Player;
let B: Player;
let C: Player;
let stranger: Player;
let cape: any;
let hat: any;
let pack: any;

test.describe.configure({ mode: 'serial' });

async function friendCall(method: string, p: Player, ids: string[]) {
  const r = await fetch(API_URL + '/v2/friend?' + ids.map((id) => 'ids=' + encodeURIComponent(id)).join('&'), { method, headers: { Authorization: 'Bearer ' + p.token } });
  expect(r.status).toBe(200);
}

async function befriend(a: Player, b: Player) {
  await friendCall('POST', a, [b.id]);
  await friendCall('POST', b, [a.id]);
}

function restoreSettings(row: string) {
  if (!row) {
    sql('DELETE FROM economy_friend_gift_settings');
    return;
  }
  const i = row.lastIndexOf('|');
  const q = (v: string) => "'" + v.replace(/'/g, "''") + "'";
  sql('UPDATE economy_friend_gift_settings SET settings=' + q(row.slice(0, i)) + '::jsonb, updated_at=' + q(row.slice(i + 1)) + '::timestamptz');
}

async function settings(change: Record<string, unknown>) {
  const cur = (await eco.get('/friend-gift-settings')).settings;
  return (await eco.put('/friend-gift-settings', { ...cur, ...change })).settings;
}

function owns(user: string, itemId: string) {
  return sql(`SELECT count(*) FROM economy_user_item WHERE user_id='${user}' AND item_id='${itemId}'`) === '1';
}

function coins(user: string) {
  return Number(sql(`SELECT COALESCE((wallet->>'coin')::bigint, 0) FROM users WHERE id='${user}'`));
}

async function send(p: Player, body: Record<string, unknown>) {
  return rpc(p, 'economy_friend_gift_send', body);
}

function code(res: { status: number; data: any }) {
  return typeof res.data === 'object' && res.data ? res.data.message || res.data.error : String(res.data);
}

test.beforeAll(async () => {
  savedRow = sql(`SELECT settings::text || '|' || updated_at::text FROM economy_friend_gift_settings`);
  const type = await base.post('/categories', { key: 'e2e_' + k + '_fgt', keyType: 'item_type', name: { fa: 'نوع هدیه' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  cape = await eco.post('/items', { key: 'e2e_' + k + '_cape', typeId: type.id, name: { fa: 'شنل', en: 'Cape' }, rarity: 'rare' });
  hat = await eco.post('/items', { key: 'e2e_' + k + '_hat', typeId: type.id, name: { fa: 'کلاه', en: 'Hat' }, giftable: false });
  made.items.push(cape.id, hat.id);
  const prize = await eco.post('/prizes', { key: 'e2e_' + k + '_fgp', name: { fa: 'جایزهٔ هدیه' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 40 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  pack = await eco.post('/products', { key: 'e2e_' + k + '_fgpack', isActive: true, name: { fa: 'بستهٔ هدیه', en: 'Gift pack' }, prizeId: prize.id, paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: 25 } });
  made.products.push(pack.id);
  A = await devicePlayer('e2e-fga-' + k, ('fga' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-fgb-' + k, ('fgb' + k).replace(/_/g, '').slice(0, 20));
  C = await devicePlayer('e2e-fgc-' + k);
  stranger = await devicePlayer('e2e-fgs-' + k);
  made.users.push(A.id, B.id, C.id, stranger.id);
  await updateAccount(B, { lang_tag: 'en' });
  await befriend(A, B);
  await befriend(A, C);
  await eco.post('/users/' + A.id + '/items', { itemId: cape.id });
  await eco.post('/users/' + A.id + '/items', { itemId: hat.id });
  for (const p of [A, B, C, stranger]) await rpc(p, 'economy_balances', {});
  sql(`UPDATE users SET wallet = jsonb_set(COALESCE(wallet,'{}'::jsonb), '{coin}', '500') WHERE id='${A.id}'`);
  await settings({ enabled: true, recipients: 'friends', minFriendDays: 0, minMutualFriends: 0, minSenderLevel: 0, minSenderAccountDays: 0, minRecipientAccountDays: 0,
    inventoryEnabled: true, itemsDefault: true, itemTypes: [], regift: 'block', equipped: 'block', currencyEnabled: true, currencies: [{ currencyKey: 'coin', min: 1, max: 100 }],
    shopEnabled: true, productsDefault: true, methods: ['in_app', 'gateway', 'store'], perSenderDaily: 20, perRecipientDaily: 20, perPairDaily: 0,
    delivery: 'instant', messageEnabled: true, messageMaxLength: 60, messageFilter: 'rules', notifyChannel: 'in_app', socketNotify: true });
});

test.afterAll(async () => {
  restoreSettings(savedRow);
  const users = made.users.map((x) => `'${x}'`).join(',');
  sql(`DELETE FROM economy_friend_gift WHERE sender_id IN (${users}) OR recipient_id IN (${users})`);
  for (const id of made.users) await deleteUser(id);
  const prods = made.products.map((x) => `'${x}'`).join(',');
  sql(`DELETE FROM economy_payment WHERE product_id IN (${prods}); DELETE FROM economy_user_prize WHERE prize_id IN (${made.prizes.map((x) => `'${x}'`).join(',')})`);
  sql(`DELETE FROM economy_user_item WHERE item_id IN (${made.items.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.products) await eco.del('/products/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('friend gifts over RPC', () => {
  test('options list what can be given, mark what the friend has and the daily limits', async () => {
    const res = await rpc(A, 'economy_friend_gift_options', { friendId: B.id, lang: 'en' });
    expect(res.status).toBe(200);
    expect(res.data).toMatchObject({ canSend: true, delivery: 'instant', friend: { id: B.id, isFriend: true }, limits: { sender: { used: 0, limit: 20 } } });
    const items = Object.fromEntries(res.data.items.map((x: any) => [x.key, x]));
    expect(items['e2e_' + k + '_cape']).toMatchObject({ giftable: true, recipientOwns: false, reason: '' });
    expect(items['e2e_' + k + '_hat']).toMatchObject({ giftable: false, reason: 'ITEM_NOT_GIFTABLE' });
    expect(res.data.currencies.find((c: any) => c.key === 'coin')).toMatchObject({ min: 1, max: 100, balance: 500 });
    expect(res.data.products.find((p: any) => p.key === 'e2e_' + k + '_fgpack')).toMatchObject({ giftable: true, paymentMethods: ['in_app'] });
    const none = await rpc(A, 'economy_friend_gift_options', { friendId: stranger.id });
    expect(none.data).toMatchObject({ canSend: false, reason: 'NOT_FRIENDS' });
  });

  test('an item moves to the friend at once, with a message, events and an inbox notice', async () => {
    const res = await send(A, { friendId: B.id, itemKey: 'e2e_' + k + '_cape', message: 'happy birthday', requestId: 'r-' + k });
    expect(res.status).toBe(200);
    expect(res.data.gift).toMatchObject({ status: 'delivered', kind: 'item', box: 'sent', message: 'happy birthday', to: { id: B.id } });
    expect(owns(A.id, cape.id)).toBe(false);
    expect(owns(B.id, cape.id)).toBe(true);
    const again = await send(A, { friendId: B.id, itemKey: 'e2e_' + k + '_cape', message: 'happy birthday', requestId: 'r-' + k });
    expect(again.data).toMatchObject({ replayed: true, gift: { id: res.data.gift.id } });
    expect(sql(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='Gift:Received' AND e.user_id='${B.id}'`)).toBe('1');
    await expect.poll(async () => JSON.stringify((await rpc(B, 'publication_inbox', { lang: 'en' })).data), { timeout: 30_000 }).toContain('sent you “Cape”');
    const got = await rpc(B, 'economy_friend_gifts', { box: 'received', lang: 'en' });
    expect(got.data.items[0]).toMatchObject({ name: 'Cape', box: 'received', from: { id: A.id } });
    expect(got.data.items[0].method).toBeUndefined();
  });

  test('refusals: a gift received cannot be given on, a non-giftable item, a stranger, a block and a bad amount', async () => {
    expect(code(await send(B, { friendId: A.id, itemKey: 'e2e_' + k + '_cape' }))).toBe('REGIFT_BLOCKED');
    expect(code(await send(A, { friendId: B.id, itemKey: 'e2e_' + k + '_hat' }))).toBe('ITEM_NOT_GIFTABLE');
    expect(code(await send(A, { friendId: stranger.id, currencyKey: 'coin', amount: 5 }))).toBe('NOT_FRIENDS');
    expect(code(await send(A, { friendId: B.id, currencyKey: 'coin', amount: 500 }))).toBe('INVALID_AMOUNT');
    expect(code(await send(A, { friendId: A.id, currencyKey: 'coin', amount: 5 }))).toBe('GIFT_SELF');
    await friendCall('POST', C, [A.id]);
    sql(`UPDATE user_edge SET state=3 WHERE source_id='${C.id}' AND destination_id='${A.id}'`);
    expect(code(await send(A, { friendId: C.id, currencyKey: 'coin', amount: 5 }))).toBe('GIFT_BLOCKED');
    sql(`UPDATE user_edge SET state=0 WHERE source_id='${C.id}' AND destination_id='${A.id}'`);
  });

  test('accept mode: currency waits, declining gives it back, accepting delivers, the daily limit stops the next one', async () => {
    await settings({ delivery: 'accept', expiryHours: 24, perPairDaily: 0, perSenderDaily: 3 });
    const before = coins(A.id);
    const g = (await send(A, { friendId: C.id, currencyKey: 'coin', amount: 30 })).data.gift;
    expect(g).toMatchObject({ status: 'pending', kind: 'currency', amount: 30 });
    expect(g.secondsLeft).toBeGreaterThan(80000);
    expect(coins(A.id)).toBe(before - 30);
    const waiting = await rpc(C, 'economy_friend_gifts', { status: 'pending' });
    expect(waiting.data.pendingCount).toBe(1);
    const dec = await rpc(C, 'economy_friend_gift_decline', { giftId: g.id });
    expect(dec.data.gift).toMatchObject({ status: 'declined', outcome: 'returned' });
    expect(coins(A.id)).toBe(before);
    const cBefore = coins(C.id);
    const g2 = (await send(A, { friendId: C.id, currencyKey: 'coin', amount: 20 })).data.gift;
    expect(code(await rpc(A, 'economy_friend_gift_accept', { giftId: g2.id }))).toBe('GIFT_NOT_FOUND');
    const acc = await rpc(C, 'economy_friend_gift_accept', { giftId: g2.id });
    expect(acc.data.gift).toMatchObject({ status: 'delivered' });
    expect(coins(C.id)).toBe(cBefore + 20);
    const sent = await rpc(A, 'economy_friend_gifts', { box: 'sent' });
    expect(sent.data.items.length).toBeGreaterThanOrEqual(3);
    expect(code(await send(A, { friendId: C.id, currencyKey: 'coin', amount: 5 }))).toBe('GIFT_SENDER_DAILY_LIMIT');
  });

  test('a shop product bought with coins for a friend reaches them; the sender pays', async () => {
    await settings({ delivery: 'instant', perSenderDaily: 0 });
    const before = coins(A.id);
    const bBefore = coins(B.id);
    const res = await send(A, { friendId: B.id, productId: pack.id, method: 'in_app' });
    expect(res.status).toBe(200);
    expect(res.data).toMatchObject({ status: 'success', gift: { status: 'delivered', kind: 'product', source: 'shop', method: 'in_app' } });
    expect(coins(A.id)).toBe(before - 25);
    expect(coins(B.id)).toBe(bBefore + 40);
    const st = await rpc(A, 'economy_purchase_status', { paymentId: res.data.paymentId });
    expect(st.data).toMatchObject({ status: 'success', friendGiftId: res.data.gift.id, friendGiftStatus: 'delivered' });
    expect(code(await send(A, { friendId: B.id, productId: pack.id, method: 'zibal' }))).toBe('METHOD_NOT_ALLOWED');
    const log = await eco.get('/friend-gifts?userId=' + A.id + '&kind=product');
    expect(log.items[0]).toMatchObject({ coinAmount: 25, paymentStatus: 'success', status: 'delivered' });
  });
});
