import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { Chat, history, notice } from '../../lib/chat';
import { sql } from '../../lib/db';

const k = uid();
const made = { users: [] as string[], products: [] as string[], prizes: [] as string[], metas: [] as string[], tags: [] as string[], rules: [] as string[] };
const sockets: Chat[] = [];
let savedRow = '';
let packKey = '';
let product: any;
let A: Player;
let B: Player;
let C: Player;

test.describe.configure({ mode: 'serial' });

function restoreSettings(row: string) {
  const i = row.lastIndexOf('|');
  const q = (v: string) => "'" + v.replace(/'/g, "''") + "'";
  sql('UPDATE base_chat_quota_settings SET settings=' + q(row.slice(0, i)) + '::jsonb, updated_at=' + q(row.slice(i + 1)) + '::timestamptz WHERE id=1');
}

async function settings(patch: Record<string, unknown>) {
  return base.put('/chat-quota/settings', { enabled: true, dailyMessages: 2, packMeta: packKey, packExpiryDays: 0, notice: 'inbox', ...patch });
}

async function socket(p: Player) {
  const c = await Chat.open(p);
  sockets.push(c);
  return c;
}

async function player(name: string, lang: string) {
  const p = await devicePlayer('e2e-cq-' + name + '-' + k, ('cq' + name + k).replace(/_/g, '').slice(0, 24));
  made.users.push(p.id);
  await updateAccount(p, { lang_tag: lang });
  return p;
}

test.beforeAll(async () => {
  savedRow = sql(`SELECT settings::text || '|' || updated_at::text FROM base_chat_quota_settings WHERE id=1`);
  packKey = 'e2e_' + k + '_dmpack';
  const meta = await eco.post('/metas', { key: packKey, name: { fa: 'بستهٔ پیام', en: 'Message pack' }, defaultValue: 0, serverOnly: false, settings: {} });
  made.metas.push(meta.id);
  const prize = await eco.post('/prizes', { key: 'e2e_' + k + '_dmprize', name: { fa: '۵ پیام' }, isRepeatable: true, items: [{ type: 'meta', key: packKey, action: 'add', value: 5 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  product = await eco.post('/products', { key: 'e2e_' + k + '_dm5', isActive: true, name: { fa: '۵ پیام خصوصی', en: '5 private messages' }, prizeId: prize.id, paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: 10 } });
  made.products.push(product.id);
  const coins = await eco.post('/prizes', { key: 'e2e_' + k + '_coins', name: { fa: 'سکه' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 100 }], tagIds: [], metadata: {} });
  made.prizes.push(coins.id);
  A = await player('a', 'fa');
  B = await player('b', 'fa');
  C = await player('c', 'en');
  await eco.post('/users/' + A.id + '/grant', { prizeId: coins.id });
  await settings({});
});

test.afterAll(async () => {
  for (const s of sockets) s.close();
  if (savedRow) restoreSettings(savedRow);
  for (const id of made.rules) await base.del('/chat-quota/rules/' + id);
  for (const id of made.users) await deleteUser(id);
  const prods = made.products.map((x) => `'${x}'`).join(',');
  if (prods) sql(`DELETE FROM economy_payment_log WHERE payment_id IN (SELECT id FROM economy_payment WHERE product_id IN (${prods})); DELETE FROM economy_payment WHERE product_id IN (${prods}); DELETE FROM economy_user_prize WHERE source_type='payment' AND prize_id IN (SELECT prize_id FROM economy_product WHERE id IN (${prods}))`);
  for (const id of made.products) await eco.del('/products/' + id);
  if (made.prizes.length) sql(`DELETE FROM economy_user_prize WHERE prize_id IN (${made.prizes.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
  sql(`DELETE FROM notification WHERE code=1120 AND user_id IN (${made.users.map((x) => `'${x}'`).join(',') || 'NULL'})`);
});

test.describe('game: private chat quota over the socket and RPC', () => {
  test('direct messages count against the free daily quota, the refusal says why with the reset time and the pack products, rooms stay free', async () => {
    const a = await socket(A);
    const dm = await a.joinDirect(B.id);
    expect(dm.startsWith('4.')).toBe(true);
    expect(await a.send('سلام ' + k)).toMatchObject({ ok: true });
    expect(await a.send('خوبی؟')).toMatchObject({ ok: true });
    const third = await a.send('سومی');
    expect(third.ok).toBe(false);
    const n = notice((third as any).message);
    expect(n).toMatchObject({ code: 'DM_QUOTA_EXCEEDED', unit: 'messages', lang: 'fa', limit: 2, used: 2, remaining: 0, packsLeft: 0, friend: false, rule: null });
    expect(n.productIds).toEqual([product.id]);
    expect(n.products[0]).toMatchObject({ id: product.id, key: 'e2e_' + k + '_dm5', name: '۵ پیام خصوصی', amount: 5 });
    expect(n.message).toContain('پیام‌های خصوصی امروزت تمام شد');
    expect(Date.parse(n.resetsAt)).toBeGreaterThan(Date.now());
    expect(Date.parse(n.resetsAt) - Date.now()).toBeLessThanOrEqual(24 * 3600 * 1000);
    expect((await a.send('چهارمی')).ok).toBe(false);
    const seen = await history(B, dm);
    expect(seen.map((m) => m.text)).toEqual(['خوبی؟', 'سلام ' + k]);
    expect(sql(`SELECT count(*) FROM notification WHERE user_id='${A.id}' AND code=1120`)).toBe('1');

    const me = await rpc(A, 'base_chat_quota', { lang: 'fa' });
    expect(me.status).toBe(200);
    expect(me.data).toMatchObject({ enabled: true, countMode: 'messages', unlimited: false, freeLeft: 0, packsLeft: 0, messages: { limit: 2, used: 2, left: 0 }, today: { sent: 2, blocked: 2 }, productIds: [product.id], timeZone: 'Asia/Tehran', rule: null });

    const room = await socket(A);
    await room.join('e2e-cq-' + k.replace(/_/g, '-').toLowerCase());
    expect(await room.send('اتاق عمومی')).toMatchObject({ ok: true });

    const c = await socket(C);
    await c.joinDirect(A.id);
    expect(await c.send('hi')).toMatchObject({ ok: true });
    expect(await c.send('hi again')).toMatchObject({ ok: true });
    const en = notice(((await c.send('third')) as any).message);
    expect(en).toMatchObject({ code: 'DM_QUOTA_EXCEEDED', lang: 'en' });
    expect(en.message).toContain("today's private messages");
  });

  test('a pack bought in the shop with coins adds messages that are spent after the free quota', async () => {
    const buy = await rpc(A, 'economy_purchase_start', { productId: product.id, method: 'in_app' });
    expect(buy.status).toBe(200);
    const me = await rpc(A, 'base_chat_quota', {});
    expect(me.data).toMatchObject({ freeLeft: 0, packsLeft: 5 });
    expect(sql(`SELECT um.value::int FROM economy_user_meta um JOIN economy_meta m ON m.id=um.meta_id WHERE m.key='${packKey}' AND um.user_id='${A.id}'`)).toBe('0');
    const a = await socket(A);
    await a.joinDirect(B.id);
    for (let i = 0; i < 5; i++) expect(await a.send('بسته ' + i)).toMatchObject({ ok: true });
    const out = notice(((await a.send('ششمی')) as any).message);
    expect(out).toMatchObject({ code: 'DM_QUOTA_EXCEEDED', packsLeft: 0 });
    const after = await rpc(A, 'base_chat_quota', {});
    expect(after.data.today).toMatchObject({ sent: 7, fromPacks: 5 });
    const report = await base.get('/chat-quota/report?days=1');
    expect(report.packs.sold).toBeGreaterThanOrEqual(1);
    expect(report.packs.products.find((p: any) => p.productId === product.id)).toMatchObject({ count: 1, coins: { coin: 10 } });
  });

  test('replies in a chat the other player started are free when the admin turns it on; parallel sends never pass the quota', async () => {
    await settings({ dailyMessages: 0, replyFree: 'started' });
    const b = await socket(B);
    await b.joinDirect(A.id);
    expect(await b.send('جواب رایگان')).toMatchObject({ ok: true });
    expect(await b.send('باز هم رایگان')).toMatchObject({ ok: true });
    const b2 = await socket(B);
    await b2.joinDirect(C.id);
    expect(notice(((await b2.send('شروع تازه')) as any).message)).toMatchObject({ code: 'DM_QUOTA_EXCEEDED', limit: 0 });

    await settings({ dailyMessages: 3 });
    const c = await socket(C);
    await c.joinDirect(B.id);
    sql(`DELETE FROM base_chat_quota_day WHERE user_id='${C.id}'`);
    const results = await Promise.all(Array.from({ length: 8 }, (_, i) => c.send('موازی ' + i)));
    expect(results.filter((r) => r.ok).length).toBe(3);
  });

  test('a tag rule makes a player unlimited; the RPC says which rule applies', async () => {
    await settings({ dailyMessages: 1 });
    const tag = await base.post('/tags', { key: 'e2e_' + k + '_vip', keyType: 'user', name: { fa: 'ویژهٔ گفتگو' }, description: {}, isPublic: false, metadata: {}, sort: 0 });
    made.tags.push(tag.id);
    const rule = await base.post('/chat-quota/rules', { name: 'e2e VIP ' + k, tagIds: [tag.id], unlimited: true, priority: 10 });
    made.rules.push(rule.id);
    await base.put('/players/' + A.id + '/tags', { tagIds: [tag.id] });
    const a = await socket(A);
    await a.joinDirect(C.id);
    for (let i = 0; i < 6; i++) expect(await a.send('ویژه ' + i)).toMatchObject({ ok: true });
    const me = await rpc(A, 'base_chat_quota', {});
    expect(me.data).toMatchObject({ unlimited: true, unlimitedBy: 'rule', rule: { id: rule.id, name: 'e2e VIP ' + k } });
    expect(me.data.today.unlimited).toBeGreaterThanOrEqual(6);
  });

  test('turning the quota off lets everyone chat again', async () => {
    await settings({ enabled: false });
    const c = await socket(C);
    await c.joinDirect(A.id);
    for (let i = 0; i < 4; i++) expect(await c.send('off ' + i)).toMatchObject({ ok: true });
    expect((await rpc(C, 'base_chat_quota', {})).data.enabled).toBe(false);
  });
});
