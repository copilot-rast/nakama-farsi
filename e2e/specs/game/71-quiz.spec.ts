import { expect, test } from '../../lib/fixtures';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { qz, removeQuiz, removeQuizPlayer } from '../../lib/quiz';
import { sql } from '../../lib/db';

function errorOf(res: { data: any }): { code: string; meta?: any; message?: string } {
  const raw = res.data && res.data.message;
  try {
    const env = JSON.parse(raw);
    if (env && typeof env === 'object') return env;
  } catch {
    return { code: raw };
  }
  return { code: raw };
}

const k = uid().replace(/_/g, '');
const users: string[] = [];
const F: Record<string, any> = {};
let rawSettings = '';

test.describe.configure({ mode: 'serial' });

async function player(tag: string): Promise<Player> {
  const p = await devicePlayer(`e2e-qz-${tag}-${k}`);
  users.push(p.id);
  return p;
}

const wallet = (p: Player, cur: string) => Number(sql(`SELECT COALESCE((wallet->>'${cur}')::bigint, 0) FROM users WHERE id='${p.id}'`).trim() || 0);
const give = (p: Player, cur: string, n: number) => sql(`UPDATE users SET wallet = jsonb_set(COALESCE(wallet::jsonb,'{}'::jsonb), '{${cur}}', to_jsonb(${n})) WHERE id='${p.id}'`);

async function ok(p: Player, id: string, body: any) {
  const r = await rpc(p, id, body);
  expect(r.status, id + ' ' + JSON.stringify(r.data)).toBe(200);
  return r.data;
}

async function answerAll(p: Player, view: any, right = true) {
  let last: any = null;
  for (const q of view.questions) {
    const key = F.answers[q.id];
    const body = q.type === 'text' ? { text: right ? key : 'nope' } : { optionId: right ? key : F.wrong[q.id] };
    last = await ok(p, 'quiz_answer', { sessionId: view.sessionId, questionId: q.id, ...body });
  }
  return last;
}

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM quiz_settings WHERE id=1`).trim();
  F.cur = 'e2e' + k + 'g';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ بازی ' + k, en: 'Game coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  await qz.setSettings((s) => { s.play.shuffleQuestions = false; s.play.shuffleOptions = false; s.timing.graceSeconds = 1; });
  F.mission = await qz.post('/missions', { key: 'e2e_' + k + '_m', name: { fa: 'ماموریت ' + k }, orderMode: 'sequential', reward: { currencies: { [F.cur]: 300 } } });
  F.p1 = (await qz.post('/packs', { key: 'e2e_' + k + '_p1', name: { fa: 'پک یک ' + k, en: 'Pack one' }, missionId: F.mission.id, sort: 1, passCorrect: 2,
    reward: { currencies: { [F.cur]: 100 } } })).pack;
  F.p2 = (await qz.post('/packs', { key: 'e2e_' + k + '_p2', name: { fa: 'پک دو ' + k }, missionId: F.mission.id, sort: 2, passPercent: 0, questionSeconds: 5 })).pack;
  F.answers = {};
  F.wrong = {};
  const opt = async (packId: string, fa: string) => {
    const q = (await qz.post('/questions', { packId, type: 'options', question: { fa }, explanation: { fa: 'توضیح ' + fa },
      options: [{ id: 0, text: { fa: 'الف' } }, { id: 1, text: { fa: 'ب' }, correct: true }, { id: 2, text: { fa: 'ج' } }] })).question;
    F.answers[q.id] = 1;
    F.wrong[q.id] = 0;
    return q;
  };
  F.q1 = await opt(F.p1.id, 'سوال یک');
  F.q2 = (await qz.post('/questions', { packId: F.p1.id, type: 'text', question: { fa: 'بزرگ‌ترین اقیانوس؟' }, answers: { fa: ['اقیانوس آرام'], en: ['Pacific'] } })).question;
  F.answers[F.q2.id] = 'اقيانوس  آرام';
  F.q3 = await opt(F.p2.id, 'سوال سه');
  F.hint = await qz.post('/questions/' + F.q1.id + '/hints', { name: { fa: 'راهنما' }, text: { fa: 'گزینهٔ دوم' }, price: { currencyKey: F.cur, amount: 30 } });
});

test.afterAll(async () => {
  for (const id of users) { removeQuizPlayer(id); await deleteUser(id); }
  removeQuiz('e2e_' + k);
  if (rawSettings) sql(`UPDATE quiz_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
  await qz.raw('GET', '/settings');
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('quiz RPCs', () => {
  test('packs and a session never carry right answers; answers are resend-safe; a pass pays the pack and the mission once', async () => {
    const p = await player('a');
    const packs = (await ok(p, 'quiz_packs', { missionKey: F.mission.key })).packs;
    expect(packs.map((x: any) => [x.key, x.locked?.code ?? null])).toEqual([[F.p1.key, null], [F.p2.key, 'QUIZ_MISSION_PACK_LOCKED']]);
    expect(packs[0]).toMatchObject({ questionCount: 2, passCorrect: 2, reward: { currencies: [{ currencyKey: F.cur, amount: 100 }] } });
    const locked = await rpc(p, 'quiz_play', { packKey: F.p2.key });
    expect(locked.status).toBe(403);
    expect(errorOf(locked)).toMatchObject({ code: 'QUIZ_MISSION_PACK_LOCKED', meta: { packKey: F.p1.key } });
    const missing = await rpc(p, 'quiz_play', { packKey: 'nope_' + k });
    expect(errorOf(missing).code).toBe('QUIZ_PACK_NOT_FOUND');

    const view = await ok(p, 'quiz_play', { packKey: F.p1.key, lang: 'fa' });
    expect(view.questions.map((q: any) => q.id)).toEqual([F.q1.id, F.q2.id]);
    expect(view.questions[0].options.map((o: any) => Object.keys(o).sort())).toEqual([['id', 'mediaUrl', 'text'], ['id', 'mediaUrl', 'text'], ['id', 'mediaUrl', 'text']]);
    expect(JSON.stringify(view)).not.toContain('اقیانوس آرام');
    expect(view.questions[0].hints).toEqual([{ id: F.hint.id, name: 'راهنما', price: { currencyKey: F.cur, amount: 30 }, owned: false }]);
    const again = await ok(p, 'quiz_play', { packKey: F.p1.key });
    expect(again).toMatchObject({ sessionId: view.sessionId, resumed: true });

    const bad = await rpc(p, 'quiz_answer', { sessionId: view.sessionId, questionId: F.q1.id, optionId: 9 });
    expect(errorOf(bad).code).toBe('QUIZ_OPTION_NOT_FOUND');
    const a1 = await ok(p, 'quiz_answer', { sessionId: view.sessionId, questionId: F.q1.id, optionId: 1 });
    expect(a1).toMatchObject({ outcome: 'correct', correctOptionIds: [1], explanation: 'توضیح سوال یک', replayed: false });
    const resend = await ok(p, 'quiz_answer', { sessionId: view.sessionId, questionId: F.q1.id, optionId: 0 });
    expect(resend).toMatchObject({ outcome: 'correct', replayed: true, correctCount: 1 });
    const a2 = await ok(p, 'quiz_answer', { sessionId: view.sessionId, questionId: F.q2.id, text: 'اقيانوس  آرام' });
    expect(a2.outcome).toBe('correct');
    expect(a2.result).toMatchObject({ passed: true, score: 100, correct: 2, total: 2, rewards: [{ currencyKey: F.cur, amount: 100 }], missionRewards: [], replayed: false });
    expect(wallet(p, F.cur)).toBe(100);
    const fin = await ok(p, 'quiz_finish', { sessionId: view.sessionId });
    expect(fin.replayed).toBe(true);

    const v2 = await ok(p, 'quiz_play', { packKey: F.p2.key });
    const r2 = await answerAll(p, v2);
    expect(r2.result).toMatchObject({ passed: true, missionsCompleted: [F.mission.key], missionRewards: [{ currencyKey: F.cur, amount: 300 }] });
    expect(wallet(p, F.cur)).toBe(400);
    const replay = await answerAll(p, await ok(p, 'quiz_play', { packKey: F.p1.key }));
    expect(replay.result).toMatchObject({ passed: true, rewards: [], missionRewards: [] });
    expect(wallet(p, F.cur)).toBe(400);
    const ms = await ok(p, 'quiz_missions', { key: F.mission.key });
    expect(ms).toMatchObject({ rewarded: true, items: [{ key: F.p1.key, passed: true }, { key: F.p2.key, passed: true }] });
    const stats = await ok(p, 'quiz_stats', {});
    expect(stats).toMatchObject({ packsPassed: 2 });
    expect(Number(sql(`SELECT count(*) FROM wallet_ledger WHERE user_id='${p.id}' AND metadata->>'sourceType'='quiz'`).trim())).toBe(2);
  });

  test('the server clock: a late answer counts as late; hints are paid once', async () => {
    const p = await player('b');
    sql(`INSERT INTO quiz_user_pack (user_id, pack_id, plays, wins, first_pass_at) VALUES ('${p.id}', '${F.p1.id}', 1, 1, now())`);
    const v = await ok(p, 'quiz_play', { packKey: F.p2.key });
    sql(`UPDATE quiz_session SET started_at = started_at - interval '20 seconds', served_at = served_at - interval '20 seconds' WHERE id='${v.sessionId}'`);
    const late = await ok(p, 'quiz_answer', { sessionId: v.sessionId, questionId: F.q3.id, optionId: 1 });
    expect(late).toMatchObject({ outcome: 'late', correct: false });
    expect(late.result).toMatchObject({ passed: true, correct: 0, missionsCompleted: [F.mission.key] });
    give(p, F.cur, 0);

    const poor = await rpc(p, 'quiz_hint_buy', { hintId: F.hint.id });
    expect(poor.status).toBe(400);
    expect(errorOf(poor).code).toBe('QUIZ_INSUFFICIENT_CURRENCY_FOR_HINT');
    give(p, F.cur, 100);
    const [b1, b2] = await Promise.all([rpc(p, 'quiz_hint_buy', { hintId: F.hint.id }), rpc(p, 'quiz_hint_buy', { hintId: F.hint.id })]);
    expect([b1.status, b2.status]).toEqual([200, 200]);
    expect(b1.data.text).toBe('گزینهٔ دوم');
    expect(wallet(p, F.cur)).toBe(70);
    const again = await ok(p, 'quiz_hint_buy', { hintId: F.hint.id });
    expect(again.replayed).toBe(true);
    expect(wallet(p, F.cur)).toBe(70);
  });

  test('question reports: reasons, only seen questions, one open report per question', async () => {
    const p = await player('d');
    const reasons = await ok(p, 'quiz_report_reasons', { lang: 'fa' });
    expect(reasons.reasons.map((r: any) => r.key)).toContain('wrong_answer');
    const unseen = await rpc(p, 'quiz_report_question', { questionId: F.q1.id, reasonKey: 'typo' });
    expect(errorOf(unseen).code).toBe('QUIZ_REPORT_NOT_SEEN');
    await ok(p, 'quiz_play', { packKey: F.p1.key });
    const r1 = await ok(p, 'quiz_report_question', { questionId: F.q1.id, reasonKey: 'typo', note: 'غلط تایپی' });
    const r2 = await ok(p, 'quiz_report_question', { questionId: F.q1.id, reasonKey: 'unclear' });
    expect(r1.duplicate).toBe(false);
    expect(r2).toEqual({ reportId: r1.reportId, duplicate: true });
    const badReason = await rpc(p, 'quiz_report_question', { questionId: F.q2.id, reasonKey: 'nope' });
    expect(errorOf(badReason).code).toBe('QUIZ_REPORT_REASON_NOT_FOUND');
  });

  test('one by one delivery sends only the current question; a daily play limit is counted on the server', async () => {
    await qz.setSettings((s) => { s.play.delivery = 'one_by_one'; s.reveal.mode = 'after_finish'; });
    const p = await player('e');
    const v = await ok(p, 'quiz_play', { packKey: F.p1.key });
    expect(v.questions.map((q: any) => q.id)).toEqual([F.q1.id]);
    const early = await rpc(p, 'quiz_answer', { sessionId: v.sessionId, questionId: F.q2.id, text: 'Pacific' });
    expect(errorOf(early).code).toBe('QUIZ_QUESTION_NOT_CURRENT');
    const a = await ok(p, 'quiz_answer', { sessionId: v.sessionId, questionId: F.q1.id, optionId: 1 });
    expect(a.correctOptionIds).toBeUndefined();
    expect(a.next.id).toBe(F.q2.id);
    const b = await ok(p, 'quiz_answer', { sessionId: v.sessionId, questionId: F.q2.id, text: 'pacific' });
    expect(b.result.review.map((x: any) => x.outcome)).toEqual(['correct', 'correct']);

    await qz.setSettings((s) => { s.play.delivery = 'all'; s.reveal.mode = 'after_answer'; s.limits.dailyPlays = 1; });
    const limited = await rpc(p, 'quiz_play', { packKey: F.p2.key });
    expect(limited.status).toBe(429);
    expect(errorOf(limited)).toMatchObject({ code: 'QUIZ_DAILY_PLAY_LIMIT', meta: { limit: 1 } });
    await qz.setSettings((s) => { s.limits.dailyPlays = 0; });
  });
});
