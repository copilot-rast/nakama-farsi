import { expect, test } from '../../lib/fixtures';
import { deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';

const k = uid();
const made = { users: [] as string[], rels: [] as string[], metaRules: [] as string[], types: [] as string[], metas: [] as string[], prizes: [] as string[] };
const M: Record<string, any> = {};
let prize: any;
let P: Player;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async () => {
  for (const [name, def, serverOnly] of [['xp', 0, false], ['lvl', 1, false], ['secret', 0, true]] as const) {
    M[name] = await eco.post('/metas', { key: 'e2e_' + k + '_' + name, name: { fa: name + ' ' + k.slice(-4), en: name }, defaultValue: def, serverOnly, settings: {} });
    made.metas.push(M[name].id);
  }
  prize = await eco.post('/prizes', { key: 'e2e_' + k + '_rp', name: { fa: 'جایزهٔ رابطه', en: 'Relation prize' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 9 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  const et = await eco.post('/event-types', { key: 'e2e_' + k + '_round', name: { fa: 'دور ' + k, en: 'Round' }, clientAllowed: true, isActive: true, rateLimit: 100, sort: 0 });
  made.types.push(et.id);
  const mr = await eco.post('/meta-rules', { key: 'e2e_' + k + '_mr', name: { fa: 'تجربه' }, eventTypeId: et.id, metaId: M.xp.id, action: 'add', valueSource: 'event', amount: 1, isActive: true });
  made.metaRules.push(mr.id);
  const up = await eco.post('/meta-relations', { key: 'e2e_' + k + '_up', name: { fa: 'سطح بالا', en: 'Level up' }, description: { fa: 'با ۱۰۰ تجربه' }, metaId: M.xp.id, operator: 'gte', value: 100,
    targets: [{ metaId: M.lvl.id, action: 'add', amount: 1 }, { metaId: M.xp.id, action: 'add', amount: -100 }], prizeType: 'prize', prizeId: prize.id, grantMode: 'claim', repeat: 'rearm', isActive: true });
  const hidden = await eco.post('/meta-relations', { key: 'e2e_' + k + '_hidden', name: { fa: 'پنهان' }, metaId: M.xp.id, operator: 'gte', value: 1, targets: [{ metaId: M.secret.id, action: 'set', amount: 1 }], clientVisible: false, isActive: true });
  const secret = await eco.post('/meta-relations', { key: 'e2e_' + k + '_secret', name: { fa: 'راز' }, metaId: M.secret.id, operator: 'eq', value: 1, targets: [{ metaId: M.lvl.id, action: 'add', amount: 10 }], isActive: true });
  made.rels.push(up.id, hidden.id, secret.id);
  P = await devicePlayer('e2e-relrpc-' + k, ('rr' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_reward_claim WHERE prize_id='${prize.id}'; DELETE FROM economy_user_prize WHERE prize_id='${prize.id}'`);
  for (const id of made.rels) await eco.del('/meta-relations/' + id);
  for (const id of made.metaRules) await eco.del('/meta-rules/' + id);
  for (const id of made.types) await eco.del('/event-types/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
});

test.describe('meta relations over RPC', () => {
  test('a client event fires a relation chain; the player lists and claims the relation prize once', async () => {
    const before = await rpc(P, 'economy_meta_relations', { lang: 'fa' });
    expect(before.status).toBe(200);
    const keys = before.data.items.map((x: any) => x.key);
    expect(keys).toContain('e2e_' + k + '_up');
    expect(keys).not.toContain('e2e_' + k + '_hidden');
    const secretItem = before.data.items.find((x: any) => x.key === 'e2e_' + k + '_secret');
    expect(secretItem.condition).toBeUndefined();
    const upItem = before.data.items.find((x: any) => x.key === 'e2e_' + k + '_up');
    expect(upItem).toMatchObject({ name: 'سطح بالا', description: 'با ۱۰۰ تجربه', completed: false, repeat: 'rearm', condition: { operator: 'gte', value: 100, current: 0, holds: false }, prize: { id: prize.id } });

    const ev = await rpc(P, 'economy_event_record', { key: 'e2e_' + k + '_round', value: 130, lang: 'fa' });
    expect(ev.status).toBe(200);
    const fired = ev.data.relations.map((r: any) => r.key);
    expect(fired).toEqual(expect.arrayContaining(['e2e_' + k + '_up', 'e2e_' + k + '_hidden', 'e2e_' + k + '_secret']));
    expect(ev.data.metas['e2e_' + k + '_xp']).toBe(30);
    expect(ev.data.metas['e2e_' + k + '_secret']).toBeUndefined();
    const claim = ev.data.claims.find((c: any) => c.kind === 'relation');
    expect(claim).toMatchObject({ grantMode: 'claim', status: 'ready', prize: { id: prize.id, name: 'جایزهٔ رابطه' } });
    const bal = await rpc(P, 'economy_balances', {});
    expect(bal.data.metas['e2e_' + k + '_lvl']).toBe(12);

    const view = await rpc(P, 'economy_meta_relations', { completedOnly: true, lang: 'fa' });
    expect(view.data.items.map((x: any) => x.key)).toEqual(expect.arrayContaining(['e2e_' + k + '_up', 'e2e_' + k + '_secret']));
    expect(view.data.items.map((x: any) => x.key)).not.toContain('e2e_' + k + '_hidden');
    expect(view.data.ready).toHaveLength(1);
    expect(view.data.ready[0]).toMatchObject({ claimId: claim.claimId, kind: 'relation', source: 'e2e_' + k + '_up', sourceName: 'سطح بالا' });
    const mine = await rpc(P, 'economy_rewards_mine', { status: 'ready', lang: 'fa' });
    expect(mine.data.items.find((x: any) => x.claimId === claim.claimId)).toMatchObject({ kind: 'relation', sourceName: 'سطح بالا' });

    const coins = (await rpc(P, 'economy_balances', {})).data.wallet.coin || 0;
    const got = await rpc(P, 'economy_reward_claim', { claimId: claim.claimId, lang: 'fa' });
    expect(got.status).toBe(200);
    expect(got.data).toMatchObject({ kind: 'relation', prize: { id: prize.id } });
    expect((await rpc(P, 'economy_balances', {})).data.wallet.coin).toBe(coins + 9);
    const again = await rpc(P, 'economy_reward_claim', { claimId: claim.claimId });
    expect(again.data.message).toBe('ALREADY_CLAIMED');
    const after = await rpc(P, 'economy_meta_relations', { lang: 'fa' });
    const upAfter = after.data.items.find((x: any) => x.key === 'e2e_' + k + '_up');
    expect(upAfter).toMatchObject({ completed: true, done: false, fires: 1 });
    expect(upAfter.claims[0]).toMatchObject({ status: 'claimed' });
  });

  test('a re-armed relation fires again after the value drops and rises, and another player cannot claim its prize', async () => {
    await rpc(P, 'economy_event_record', { key: 'e2e_' + k + '_round', value: 75 });
    const view = await rpc(P, 'economy_meta_relations', { completedOnly: true });
    const up = view.data.items.find((x: any) => x.key === 'e2e_' + k + '_up');
    expect(up.fires).toBe(2);
    expect(sql(`SELECT value::int FROM economy_user_meta WHERE user_id='${P.id}' AND meta_id='${M.xp.id}'`)).toBe('5');
    const open = up.claims.find((c: any) => c.status === 'claimable');
    const other = await devicePlayer('e2e-relrpc2-' + k);
    made.users.push(other.id);
    const steal = await rpc(other, 'economy_reward_claim', { claimId: open.claimId });
    expect(steal.data.message).toBe('CLAIM_NOT_FOUND');
    const evs = sql(`SELECT count(*) FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id WHERE t.key='MetaRelation:Completed' AND e.user_id='${P.id}' AND e.metadata->>'key'='e2e_${k}_up'`);
    expect(evs).toBe('2');
  });
});
