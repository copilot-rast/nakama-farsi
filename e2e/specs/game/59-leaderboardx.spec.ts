import { expect, test } from '../../lib/fixtures';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { errorOf, lbx, removeBoards } from '../../lib/leaderboardx';
import { sql } from '../../lib/db';

const k = uid().replace(/_/g, '');
const part = 'e2e_' + k;
const users: string[] = [];
const F: Record<string, any> = {};
let rawSettings = '';

test.describe.configure({ mode: 'serial' });

async function player(tag: string): Promise<Player> {
  const p = await devicePlayer(`e2e-lb-${tag}-${k}`);
  users.push(p.id);
  return p;
}

const wallet = (p: Player, cur: string) => Number(sql(`SELECT COALESCE((wallet->>'${cur}')::bigint, 0) FROM users WHERE id='${p.id}'`).trim() || 0);

async function ok(p: Player, id: string, payload: any) {
  const r = await rpc(p, id, payload);
  expect(r.status, JSON.stringify(r.data)).toBe(200);
  return r.data;
}

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM leaderboardx_settings WHERE id=1`).trim();
  await lbx.setSettings((s) => { s.events.economyEvent = ''; s.events.analyticsEvent = ''; s.rewards.payMode = 'auto'; });
  F.cur = 'e2e' + k + 'c';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ جدول ' + k, en: 'Board coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  F.ev = 'e2e_' + k + '_arrest';
  F.type = await eco.post('/event-types', { key: F.ev, name: { fa: 'دستگیری ' + k, en: 'Arrest ' + k }, description: {}, clientAllowed: true, valueMin: 0, valueMax: 100, rateLimit: 100, isActive: true, sort: 0 });
  F.a = await player('a');
  F.b = await player('b');
  F.c = await player('c');
});

test.afterAll(async () => {
  await removeBoards(part);
  for (const id of users) await deleteUser(id);
  if (rawSettings) sql(`UPDATE leaderboardx_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
  await lbx.raw('GET', '/settings');
  if (F.type) await eco.del('/event-types/' + F.type.id);
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('leaderboardx game RPCs', () => {
  test('economy events fill a weekly board: top, around me, my tier and the next tier', async () => {
    F.weekly = await lbx.board({ key: part + '_weekly', name: { fa: 'دستگیری هفته ' + k, en: 'Arrests of the week' }, period: 'weekly',
      source: { type: 'economy_event', event: F.ev, calc: 'value', multiplier: 1 },
      tierBasis: 'score', tiers: [{ key: 'gold', name: { fa: 'طلایی', en: 'Gold' }, threshold: 100 }, { key: 'silver', name: { fa: 'نقره‌ای', en: 'Silver' }, threshold: 10 }],
      rewards: [{ fromRank: 1, toRank: 1, currencies: { [F.cur]: 50 } }] });
    for (const [p, values] of [[F.a, [60, 70]], [F.b, [30]], [F.c, [5]]] as [Player, number[]][]) {
      for (const v of values) await ok(p, 'economy_event_record', { key: F.ev, value: v });
    }
    const view = await ok(F.b, 'leaderboardx_board', { key: F.weekly.key, lang: 'en', around: 3 });
    expect(view).toMatchObject({ key: F.weekly.key, name: 'Arrests of the week', status: 'active', period: 'weekly', players: 3, tierBasis: 'score' });
    expect(view.periodEnd).toBeTruthy();
    expect(view.top.map((r: any) => [r.userId, r.score, r.tier])).toEqual([[F.a.id, 130, 'gold'], [F.b.id, 30, 'silver'], [F.c.id, 5, null]]);
    expect(view.me).toMatchObject({ rank: 2, score: 30, tier: 'silver', nextTier: { key: 'gold', need: 70 } });
    expect(view.around.some((r: any) => r.me && r.userId === F.b.id)).toBe(true);
    expect(view.tiers[0]).toMatchObject({ key: 'gold', name: 'Gold', threshold: 100 });
    expect(view.rewards[0]).toMatchObject({ fromRank: 1, toRank: 1, currencies: [{ currencyKey: F.cur, amount: 50 }] });
    const list = await ok(F.c, 'leaderboardx_list', { lang: 'fa' });
    const mine = list.boards.find((b: any) => b.key === F.weekly.key);
    expect(mine).toMatchObject({ name: 'دستگیری هفته ' + k, me: { rank: 3, score: 5, tier: null, nextTier: { key: 'silver', need: 5 } } });
    const missing = await rpc(F.c, 'leaderboardx_board', { key: part + '_nope' });
    expect(errorOf(missing).code).toBe('LEADERBOARD_NOT_FOUND');
  });

  test('client submits obey the board: not allowed, request id, cap per write, resend', async () => {
    const res = await rpc(F.a, 'leaderboardx_submit', { key: F.weekly.key, score: 5, requestId: 'x1' });
    expect(errorOf(res).code).toBe('LEADERBOARD_SUBMIT_NOT_ALLOWED');
    F.race = await lbx.board({ key: part + '_race', name: { fa: 'بهترین زمان ' + k }, sortOrder: 'asc', operator: 'best',
      source: { type: 'client', calc: 'value', multiplier: 1 }, limits: { maxPerWrite: 600, perMinute: 10, onExceed: 'reject' } });
    expect(errorOf(await rpc(F.a, 'leaderboardx_submit', { key: F.race.key, score: 50 })).code).toBe('REQUEST_ID_REQUIRED');
    const high = errorOf(await rpc(F.a, 'leaderboardx_submit', { key: F.race.key, score: 900, requestId: 'r0' }));
    expect(high.code).toBe('LEADERBOARD_SCORE_TOO_HIGH');
    expect(high.meta).toMatchObject({ max: 600 });
    expect(errorOf(await rpc(F.a, 'leaderboardx_submit', { key: F.race.key, score: -3, requestId: 'r1' })).code).toBe('LEADERBOARD_INVALID_SCORE');
    expect(await ok(F.a, 'leaderboardx_submit', { key: F.race.key, score: 95, requestId: 'r2' })).toMatchObject({ accepted: true, score: 95, rank: 1, replayed: false });
    expect(await ok(F.a, 'leaderboardx_submit', { key: F.race.key, score: 120, requestId: 'r3' })).toMatchObject({ accepted: true, score: 95 });
    expect(await ok(F.a, 'leaderboardx_submit', { key: F.race.key, score: 95, requestId: 'r2' })).toMatchObject({ accepted: false, replayed: true, score: 95 });
    expect(await ok(F.b, 'leaderboardx_submit', { key: F.race.key, score: 80, requestId: 'b1' })).toMatchObject({ rank: 1 });
    const view = await ok(F.a, 'leaderboardx_board', { key: F.race.key });
    expect(view.top.map((r: any) => r.score)).toEqual([80, 95]);
    expect(view.me.rank).toBe(2);
  });

  test('claim mode: a closed period waits in the history, the claim pays once, winners are listed', async () => {
    await lbx.put('/boards/' + F.weekly.id, { ...F.weekly, payMode: 'claim' });
    await lbx.post('/boards/' + F.weekly.id + '/close', { pay: true });
    const hist = await ok(F.a, 'leaderboardx_history', { key: F.weekly.key });
    expect(hist.claimable).toBeGreaterThanOrEqual(1);
    expect(hist.items[0]).toMatchObject({ boardKey: F.weekly.key, rank: 1, tier: 'gold', score: 130, status: 'claimable', rewards: { currencies: [{ currencyKey: F.cur, amount: 50 }] } });
    expect(wallet(F.a, F.cur)).toBe(0);
    const pid = hist.items[0].periodId;
    const nothing = await rpc(F.b, 'leaderboardx_claim', { periodId: pid });
    expect(errorOf(nothing).code).toBe('LEADERBOARD_REWARD_NOT_CLAIMABLE');
    const claimed = await ok(F.a, 'leaderboardx_claim', { periodId: pid });
    expect(claimed.claimed[0]).toMatchObject({ status: 'paid', replayed: false, rank: 1 });
    expect(wallet(F.a, F.cur)).toBe(50);
    expect((await ok(F.a, 'leaderboardx_claim', { periodId: pid })).claimed[0].replayed).toBe(true);
    expect(wallet(F.a, F.cur)).toBe(50);
    const win = await ok(F.c, 'leaderboardx_winners', { key: F.weekly.key });
    expect(win.period).toMatchObject({ periodId: pid, players: 3, reason: 'manual' });
    expect(win.winners.map((w: any) => [w.rank, w.userId, w.tier])).toEqual([[1, F.a.id, 'gold'], [2, F.b.id, 'silver'], [3, F.c.id, null]]);
    const fresh = await ok(F.a, 'leaderboardx_board', { key: F.weekly.key });
    expect(fresh.top).toEqual([]);
    expect(fresh.me).toBeNull();
  });

  test('a core reset every minute closes the period by itself and pays the winner', async () => {
    test.setTimeout(200_000);
    const b = await lbx.board({ key: part + '_minute', name: { fa: 'دقیقه‌ای ' + k }, period: 'custom', resetSchedule: '* * * * *',
      source: { type: 'client', calc: 'value', multiplier: 1 }, rewards: [{ fromRank: 1, toRank: 2, currencies: { [F.cur]: 9 } }] });
    while (new Date().getSeconds() > 45) await new Promise((r) => setTimeout(r, 1000));
    await ok(F.c, 'leaderboardx_submit', { key: b.key, score: 3, requestId: 'm1' });
    await expect.poll(async () => (await ok(F.c, 'leaderboardx_history', { key: b.key })).items[0]?.status, { timeout: 170_000, intervals: [5000] }).toBe('paid');
    expect(wallet(F.c, F.cur)).toBe(9);
    const reason = sql(`SELECT reason FROM leaderboardx_period WHERE board_key='${b.key}' ORDER BY period_end LIMIT 1`).trim();
    expect(reason).toBe('reset');
  });
});
