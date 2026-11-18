import { expect, test } from '../../lib/fixtures';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { cg, removeCampaign } from '../../lib/campaign';
import { errorOf } from '../../lib/leaderboardx';
import { sql } from '../../lib/db';

const k = uid().replace(/_/g, '');
const users: string[] = [];
const F: Record<string, any> = {};
let rawSettings = '';

test.describe.configure({ mode: 'serial' });

async function player(tag: string): Promise<Player> {
  const p = await devicePlayer(`e2e-cg-${tag}-${k}`);
  users.push(p.id);
  return p;
}

const wallet = (p: Player, cur: string) => Number(sql(`SELECT COALESCE((wallet->>'${cur}')::bigint, 0) FROM users WHERE id='${p.id}'`).trim() || 0);
const metaValue = (p: Player) => Number(sql(`SELECT COALESCE((SELECT value FROM economy_user_meta um JOIN economy_meta m ON m.id=um.meta_id WHERE um.user_id='${p.id}' AND m.key='${F.xpKey}'), 0)::int`).trim() || 0);

async function start(p: Player, stageKey: string) {
  const r = await rpc(p, 'campaign_attempt_start', { stageKey });
  expect(r.status, JSON.stringify(r.data)).toBe(200);
  return r.data;
}

async function complete(p: Player, attemptId: string, body: any) {
  return rpc(p, 'campaign_attempt_complete', { attemptId, ...body });
}

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM campaign_settings WHERE id=1`).trim();
  F.cur = 'e2e' + k + 'c';
  F.xpKey = 'e2e' + k + 'xp';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ مرحله ' + k, en: 'Stage coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  F.meta = await eco.post('/metas', { key: F.xpKey, name: { fa: 'تجربهٔ ' + k, en: 'XP ' + k } });
  F.prize = await eco.post('/prizes', { key: 'e2e_' + k + '_gold', name: { fa: 'طلای ' + k }, message: {}, tagIds: [], isRepeatable: true, metadata: {}, items: [{ type: 'currency', key: F.cur, action: 'add', value: 7 }] });
  await cg.setSettings((s) => { s.xp.metaKey = F.xpKey; s.attempts.graceSeconds = 3; s.attempts.defaultMinSeconds = 0; s.stars.missingField = 'pass'; s.events.economyEvent = ''; });
  F.ch1 = await cg.post('/chapters', { key: 'e2e_' + k + '_ch1', name: { fa: 'فصل یک ' + k, en: 'Chapter one' }, isActive: true, sort: 9001,
    milestones: [{ percent: 50, currencies: { [F.cur]: 40 } }] });
  F.ch2 = await cg.post('/chapters', { key: 'e2e_' + k + '_ch2', name: { fa: 'فصل دو ' + k }, isActive: true, sort: 9002, unlock: [{ type: 'chapter', key: F.ch1.key }] });
  const stage = (key: string, chapterId: string, extra: any) => cg.post('/stages', { key: 'e2e_' + k + '_' + key, chapterId, name: { fa: 'مرحلهٔ ' + key }, isActive: true, ...extra });
  F.s1 = await stage('s1', F.ch1.id, { sort: 1, mode: 'ordered', timeLimitSeconds: 60, xp: 100, rewards: { currencies: { [F.cur]: 1000 } }, tierPrizes: [{ stars: 3, prizeKey: F.prize.key }] });
  F.s2 = await stage('s2', F.ch1.id, { sort: 2, mode: 'ordered', rewards: { currencies: { [F.cur]: 10 } } });
  F.s3 = await stage('s3', F.ch2.id, { sort: 1, mode: 'ordered' });
  F.quiz = await stage('quiz', F.ch1.id, { sort: 3, mode: 'once', rewards: { currencies: { [F.cur]: 5000 } } });
  F.ride = await stage('ride', F.ch1.id, { sort: 4, mode: 'repeatable', cooldownSeconds: 3600, dailyLimit: 3, rewardGroup: 'e2e_' + k + '_roam', rewards: { currencies: { [F.cur]: 100 } } });
  F.fr = await cg.post('/freeride', { key: 'e2e_' + k + '_bonus', name: { fa: 'پاداش گردش ' + k }, isActive: true, rewards: { currencies: { [F.cur]: 3 } }, xp: 2, cooldownSeconds: 3600, dailyLimit: 2 });
});

test.afterAll(async () => {
  for (const id of users) await deleteUser(id);
  removeCampaign('e2e_' + k);
  if (rawSettings) sql(`UPDATE campaign_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
  await cg.raw('GET', '/settings');
  if (F.prize) { sql(`DELETE FROM economy_user_prize WHERE prize_id='${F.prize.id}'`); await eco.del('/prizes/' + F.prize.id); }
  sql(`DELETE FROM economy_user_meta WHERE meta_id IN (SELECT id FROM economy_meta WHERE key='${F.xpKey}')`);
  if (F.meta) await eco.del('/metas/' + F.meta.id);
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('campaign RPCs', () => {
  test('chapters list locks; a server-timed run gets stars from the server, pays the difference and a resend pays nothing', async () => {
    const p = await player('a');
    let view = (await rpc(p, 'campaign_chapters', { lang: 'fa' })).data;
    const ch1 = view.chapters.find((c: any) => c.key === F.ch1.key);
    const ch2 = view.chapters.find((c: any) => c.key === F.ch2.key);
    expect(ch1.unlocked).toBe(true);
    expect(ch2.unlocked).toBe(false);
    expect(ch2.locks).toEqual([{ type: 'chapter', key: F.ch1.key }]);
    expect(ch1.stages.map((s: any) => [s.key, s.unlocked])).toEqual([[F.s1.key, true], [F.s2.key, false], [F.quiz.key, true], [F.ride.key, true]]);
    expect(ch1.stages[0].rewards).toEqual([{ currencyKey: F.cur, amount: 1000 }]);

    const locked = await rpc(p, 'campaign_attempt_start', { stageKey: F.s2.key });
    expect(locked.status).toBe(403);
    expect(errorOf(locked).code).toBe('STAGE_LOCKED');

    const a = await start(p, F.s1.key);
    expect(a.attemptId).toMatch(/^[0-9a-f-]{36}$/);
    const tooLong = await complete(p, a.attemptId, { success: true, elapsedSeconds: 50, healthPercent: 90 });
    expect(tooLong.status).toBe(400);
    expect(errorOf(tooLong).code).toBe('INVALID_RESULT');
    sql(`UPDATE campaign_attempt SET started_at = started_at - interval '55 seconds' WHERE id='${a.attemptId}'`);
    const done = await complete(p, a.attemptId, { success: true, elapsedSeconds: 50, healthPercent: 60, fields: { arrests: 2 } });
    expect(done.status, JSON.stringify(done.data)).toBe(200);
    expect(done.data).toMatchObject({ stars: 2, bestStars: 2, xpGained: 50, rewards: [{ currencyKey: F.cur, amount: 500 }], replayed: false, rewardLimited: null });
    expect(wallet(p, F.cur)).toBe(500);
    expect(metaValue(p)).toBe(50);
    const again = await complete(p, a.attemptId, { success: true, elapsedSeconds: 50, healthPercent: 60 });
    expect(again.data).toMatchObject({ stars: 2, replayed: true });
    expect(wallet(p, F.cur)).toBe(500);

    const b = await start(p, F.s1.key);
    sql(`UPDATE campaign_attempt SET started_at = started_at - interval '30 seconds' WHERE id='${b.attemptId}'`);
    const three = await complete(p, b.attemptId, { success: true, elapsedSeconds: 20, healthPercent: 100 });
    expect(three.data).toMatchObject({ stars: 3, xpGained: 50, rewards: [{ currencyKey: F.cur, amount: 500 }] });
    expect(three.data.prizes.map((x: any) => x.key)).toEqual([F.prize.key]);
    expect(three.data.milestoneRewards).toEqual([{ currencyKey: F.cur, amount: 40 }]);
    expect(wallet(p, F.cur)).toBe(1000 + 40 + 7);
    const ledger = Number(sql(`SELECT count(*) FROM wallet_ledger WHERE user_id='${p.id}' AND metadata->>'sourceType'='stage'`).trim());
    expect(ledger).toBeGreaterThanOrEqual(3);

    const c = await start(p, F.s1.key);
    const lower = await complete(p, c.attemptId, { success: true, elapsedSeconds: 0.5, healthPercent: 10 });
    expect(lower.data).toMatchObject({ stars: 1, bestStars: 3, rewards: [], xpGained: 0 });
    const closed = await start(p, F.s2.key);
    const d = await start(p, F.s2.key);
    expect(errorOf(await complete(p, closed.attemptId, { success: true, elapsedSeconds: 0 })).code).toBe('ATTEMPT_CLOSED');
    expect((await complete(p, d.attemptId, { success: true, elapsedSeconds: 0 })).data.stars).toBe(3);

    view = (await rpc(p, 'campaign_chapters', {})).data;
    expect(view.chapters.find((x: any) => x.key === F.ch2.key).unlocked).toBe(true);
    expect(view.xp).toBe(100);
  });

  test('a one-time stage pays once; a repeatable stage and a free ride reward share caps', async () => {
    const p = await player('b');
    for (let i = 0; i < 2; i++) {
      const a = await start(p, F.quiz.key);
      await complete(p, a.attemptId, { success: true, elapsedSeconds: 0 });
    }
    expect(wallet(p, F.cur)).toBe(5000);
    const r1 = await start(p, F.ride.key);
    expect((await complete(p, r1.attemptId, { success: true, elapsedSeconds: 1 })).data).toMatchObject({ rewardLimited: null, rewards: [{ currencyKey: F.cur, amount: 100 }] });
    const r2 = await start(p, F.ride.key);
    const limited = (await complete(p, r2.attemptId, { success: true, elapsedSeconds: 1 })).data;
    expect(limited.rewardLimited).toBe('cooldown');
    expect(limited.nextRewardAt).toBeTruthy();
    expect(wallet(p, F.cur)).toBe(5100);

    const noId = await rpc(p, 'campaign_freeride_claim', { key: F.fr.key });
    expect(errorOf(noId).code).toBe('REQUEST_ID_REQUIRED');
    const c1 = await rpc(p, 'campaign_freeride_claim', { key: F.fr.key, requestId: 'r1-' + k });
    expect(c1.data).toMatchObject({ rewards: [{ currencyKey: F.cur, amount: 3 }], xpGained: 2, replayed: false });
    expect((await rpc(p, 'campaign_freeride_claim', { key: F.fr.key, requestId: 'r1-' + k })).data.replayed).toBe(true);
    const c2 = await rpc(p, 'campaign_freeride_claim', { key: F.fr.key, requestId: 'r2-' + k });
    expect(c2.status).toBe(409);
    expect(errorOf(c2)).toMatchObject({ code: 'REWARD_LIMITED', meta: { reason: 'cooldown' } });
    expect(wallet(p, F.cur)).toBe(5103);
    const view = (await rpc(p, 'campaign_chapters', {})).data;
    const fr = view.freeride.find((x: any) => x.key === F.fr.key);
    expect(fr.nextRewardAt).toBeTruthy();
    expect(fr.rewardsLeftToday).toBe(1);
  });
});
