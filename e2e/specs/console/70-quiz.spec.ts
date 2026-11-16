import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { qz, removeQuiz, removeQuizPlayer } from '../../lib/quiz';
import { noHorizontalScroll } from '../../lib/console';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const F: Record<string, any> = {};
const users: string[] = [];
let rawSettings = '';
const MS = 'e2e_' + k + '_mission';
const PK = 'e2e_' + k + '_pack';
const IMP = 'e2e_' + k + '_imported';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#quiz-root');
const panel = (page: Page) => page.locator('[data-quiz=panel] [role=dialog]');

async function openQuiz(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/quiz/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-quiz-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM quiz_settings WHERE id=1`).trim();
  F.cur = 'e2e' + k + 'q';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ آزمون ' + k, en: 'Quiz coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
});

test.afterAll(async () => {
  for (const id of users) { removeQuizPlayer(id); await deleteUser(id); }
  removeQuiz('e2e_' + k);
  if (rawSettings) sql(`UPDATE quiz_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('quiz console', () => {
  test('sidebar group; a sequential mission and a pack with a pass rule, a reward and a question count', async ({ page }) => {
    await openQuiz(page, 'missions');
    await expect(page.locator('[data-quiz-group] a[href="#/quiz/questions"]')).toBeAttached();
    await root(page).locator('[data-quiz-action="new-mission"]').click();
    let p = panel(page);
    await p.locator('input[name="key"]').fill(MS);
    await p.locator('input[name="name-fa"]').fill('ماموریت آزمون ' + k);
    await p.locator('select[name="orderMode"]').selectOption('sequential');
    await p.locator('[data-quiz-action="add-currency-mission"]').click();
    await p.locator('[data-quiz-currencies="mission"] select[name="reward-currency"]').selectOption(F.cur);
    await p.locator('[data-quiz-currencies="mission"] input[name="reward-amount"]').fill('۲۵۰');
    await p.locator('[data-quiz-action="save-mission"]').click();
    await expect(p).toHaveCount(0);
    await expect(root(page).locator(`tr[data-quiz-row="${MS}"]`)).toContainText('به ترتیب');
    F.mission = (await qz.get('/missions')).items.find((m: any) => m.key === MS);
    expect(F.mission).toMatchObject({ kind: 'mission', orderMode: 'sequential', reward: { currencies: { [F.cur]: 250 } } });

    await openQuiz(page, 'packs');
    await root(page).locator('[data-quiz-action="new-pack"]').click();
    p = panel(page);
    await p.locator('input[name="key"]').fill('bad key!');
    await p.locator('input[name="name-fa"]').fill('پک ' + k);
    await p.locator('[data-quiz-action="save-pack"]').click();
    await expect(p.locator('.eco-notice')).toContainText('کلید درست نیست');
    await p.locator('input[name="key"]').fill(PK);
    await p.locator('input[name="name-en"]').fill('Pack ' + k);
    await p.locator('select[name="missionId"]').selectOption(F.mission.id);
    await p.locator('input[name="questionCount"]').fill('۲');
    await p.locator('input[name="passCorrect"]').fill('۱');
    await p.locator('input[name="questionSeconds"]').fill('30');
    await p.locator('input[name="category"]').fill('عمومی');
    await p.locator('[data-quiz-action="add-currency-pack"]').click();
    await p.locator('[data-quiz-currencies="pack"] select[name="reward-currency"]').selectOption(F.cur);
    await p.locator('[data-quiz-currencies="pack"] input[name="reward-amount"]').fill('100');
    await p.locator('[data-quiz-action="save-pack"]').click();
    await expect(p).toHaveCount(0);
    const row = root(page).locator(`tr[data-quiz-row="${PK}"]`);
    await expect(row).toContainText('۱ پاسخ درست');
    await expect(row).toContainText('۲ سوال در هر بار');
    const pk = (await qz.get('/packs?q=' + PK)).items[0];
    expect(pk).toMatchObject({ missionKey: MS, questionCount: 2, passCorrect: 1, passPercent: null, questionSeconds: 30, category: 'عمومی', reward: { currencies: { [F.cur]: 100 } } });
    F.pack = pk;
  });

  test('questions editor: options in fa and en, a right option is required, hints after saving', async ({ page }) => {
    await openQuiz(page, 'questions?packId=' + F.pack.id);
    await root(page).locator('[data-quiz-action="new-question"]').click();
    const p = panel(page);
    await expect(p.locator('select[name="packId"]')).toHaveValue(F.pack.id);
    await p.locator('textarea[name="question-fa"]').fill('پایتخت ایران کجاست؟ ' + k);
    await p.locator('textarea[name="question-en"]').fill('Capital of Iran? ' + k);
    const opts = p.locator('[data-quiz-option]');
    await expect(opts).toHaveCount(4);
    await opts.nth(0).locator('input[name="option-correct"]').uncheck();
    const answers = [['شیراز', 'Shiraz'], ['تهران', 'Tehran'], ['تبریز', 'Tabriz'], ['مشهد', 'Mashhad']];
    for (let i = 0; i < 4; i++) {
      await opts.nth(i).locator('input[name="option-fa"]').fill(answers[i][0]);
      await opts.nth(i).locator('input[name="option-en"]').fill(answers[i][1]);
    }
    await p.locator('[data-quiz-action="save-question"]').click();
    await expect(p.locator('.eco-notice')).toContainText('دست‌کم یک گزینه را درست علامت بزنید');
    await opts.nth(1).locator('input[name="option-correct"]').check();
    await p.locator('textarea[name="explanation-fa"]').fill('تهران پایتخت است.');
    await p.locator('[data-quiz-action="save-question"]').click();
    await expect(panel(page).locator('[data-quiz-action="add-hint"]')).toBeVisible();
    await panel(page).locator('[data-quiz-action="add-hint"]').click();
    const hf = panel(page).locator('[data-quiz-form="hint"]');
    await hf.locator('input[name="hint-name-fa"]').fill('راهنما');
    await hf.locator('textarea[name="hint-text-fa"]').fill('شهر برج میلاد');
    await hf.locator('select[name="hint-currency"]').selectOption(F.cur);
    await hf.locator('input[name="hint-amount"]').fill('20');
    await panel(page).locator('[data-quiz-action="save-hint"]').click();
    await expect(panel(page).locator('[data-quiz-hint]')).toHaveCount(1);
    await expect(panel(page).locator('[data-quiz-hint]')).toContainText('۲۰ ' + F.cur);
    const q = (await qz.get('/questions?packId=' + F.pack.id)).items[0];
    expect(q.options.map((o: any) => [o.text.en, o.correct])).toEqual([['Shiraz', false], ['Tehran', true], ['Tabriz', false], ['Mashhad', false]]);
    expect(q).toMatchObject({ type: 'options', hints: 1, explanation: { fa: 'تهران پایتخت است.' } });
    F.q1 = q;
    F.q2 = (await qz.post('/questions', { packId: F.pack.id, type: 'text', question: { fa: 'بزرگ‌ترین اقیانوس؟' }, answers: { fa: ['آرام'], en: ['Pacific'] } })).question;
  });

  test('bulk import: a check-only run, errors refuse the whole file, then the import', async ({ page }) => {
    await openQuiz(page, 'import');
    const doc = {
      packs: [{ key: IMP, name: { fa: 'پک واردشده ' + k }, questionCount: 1 }],
      questions: [{ key: IMP + '_q1', packKey: IMP, question: { fa: 'دو به علاوهٔ دو؟' }, options: [{ id: 0, text: { fa: '۳' } }, { id: 1, text: { fa: '۴' }, correct: true }] }],
      rows: [{ 'Pack Key': IMP, Question: 'Largest planet?', 'Option 1': 'Mars', 'Option 2': 'Jupiter', 'Correct Options': '2' }],
    };
    await root(page).locator('textarea[name="import-json"]').fill(JSON.stringify({ ...doc, rows: [{ 'Pack Key': IMP, Question: 'x', 'Option 1': 'a', 'Option 2': 'b', 'Correct Options': '7' }] }));
    await root(page).locator('[data-quiz-action="import"]').click();
    await expect(root(page).locator('[data-quiz-import-result]')).toContainText('فایل خطا دارد');
    await expect(root(page).locator('[data-quiz-import-result] table')).toContainText('Correct Options');
    expect(sql(`SELECT count(*) FROM quiz_pack WHERE key='${IMP}'`).trim()).toBe('0');
    await root(page).locator('textarea[name="import-json"]').fill(JSON.stringify(doc));
    await root(page).locator('[data-quiz-action="import-dry"]').click();
    await expect(root(page).locator('[data-quiz-import-result]')).toContainText('فایل درست است');
    expect(sql(`SELECT count(*) FROM quiz_pack WHERE key='${IMP}'`).trim()).toBe('0');
    await root(page).locator('[data-quiz-action="import"]').click();
    await expect(root(page).locator('[data-quiz-import-result]')).toContainText('سوال: ۲ تازه');
    expect(sql(`SELECT count(*) FROM quiz_question q JOIN quiz_pack k ON k.id=q.pack_id WHERE k.key='${IMP}'`).trim()).toBe('2');
    const exp = await qz.get('/export?packId=' + sql(`SELECT id FROM quiz_pack WHERE key='${IMP}'`).trim());
    expect(exp.questions.length).toBe(2);
  });

  test('a player report is reviewed: resolved and the question switched off; stats and the player card follow a play', async ({ page }) => {
    const pl: Player = await devicePlayer('e2e-qzc-' + k);
    users.push(pl.id);
    const play = (await rpc(pl, 'quiz_play', { packKey: PK })).data;
    expect(play.questions.length).toBe(2);
    for (const q of play.questions) {
      const body = q.type === 'text' ? { text: 'آرام' } : { optionId: 1 };
      await rpc(pl, 'quiz_answer', { sessionId: play.sessionId, questionId: q.id, ...body });
    }
    const rep = await rpc(pl, 'quiz_report_question', { questionId: F.q1.id, reasonKey: 'wrong_answer', note: 'گزینه اشتباه ' + k });
    expect(rep.status, JSON.stringify(rep.data)).toBe(200);

    await openQuiz(page, 'reports');
    const row = root(page).locator(`tr[data-quiz-row="${rep.data.reportId}"]`);
    await expect(row).toContainText('گزینه اشتباه ' + k);
    await row.locator(`[data-quiz-review="${rep.data.reportId}"]`).click();
    const p = panel(page);
    await p.locator('textarea[name="resolution"]').fill('درست شد');
    await p.locator('input[name="disableQuestion"]').check();
    await p.locator('[data-quiz-action="review-resolved"]').click();
    await expect(p).toHaveCount(0);
    expect(sql(`SELECT r.status || ':' || q.is_active FROM quiz_question_report r JOIN quiz_question q ON q.id=r.question_id WHERE r.id='${rep.data.reportId}'`).trim()).toBe('resolved:false');
    await qz.put('/questions/' + F.q1.id, { ...F.q1, isActive: true });

    await openQuiz(page, 'stats');
    await expect(root(page).locator(`tr[data-quiz-row="${PK}"]`)).toContainText('۱۰۰٪');

    await page.goto(CONSOLE_URL + '/#/players/' + pl.id);
    const card = page.locator('[data-quiz-player]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    const pr = card.locator(`[data-quiz-player-pack="${PK}"]`);
    await expect(pr).toContainText('قبول');
    await pr.locator(`[data-quiz-reset="${PK}"]`).click();
    await page.locator('[data-quiz-confirm="ok"]').click();
    await expect(card.locator(`[data-quiz-player-pack="${PK}"]`)).toContainText('هنوز قبول نشده');
    expect(sql(`SELECT plays || ':' || rewards FROM quiz_user_pack WHERE user_id='${pl.id}' AND pack_id='${F.pack.id}'`).trim()).toBe('0:1');
  });

  test('a result report with placeholders; a private webhook is refused', async ({ page }) => {
    await openQuiz(page, 'results');
    await root(page).locator('[data-quiz-action="new-template"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill('e2e_' + k + '_result');
    await p.locator('input[name="name"]').fill('نتیجه ' + k);
    await p.locator('select[name="missionId"]').selectOption(F.mission.id);
    await p.locator('textarea[name="body"]').fill('نتیجه: ');
    await p.locator('[data-quiz-placeholders] button', { hasText: '{status}' }).click();
    await p.locator('input[name="webhookUrl"]').fill('http://127.0.0.1/hook');
    await p.locator('[data-quiz-action="save-template"]').click();
    await expect(p.locator('.eco-notice')).toContainText('آدرس وب‌هوک');
    await p.locator('input[name="webhookUrl"]').fill('');
    await p.locator('[data-quiz-action="save-template"]').click();
    await expect(p).toHaveCount(0);
    const tpl = (await qz.get('/templates')).items.find((x: any) => x.key === 'e2e_' + k + '_result');
    expect(tpl.body).toBe('نتیجه: {status}');
  });

  test('settings: reveal and hint pricing are saved; a bad time zone is refused; pages fit 390px in English dark mode', async ({ page, spawn }) => {
    await openQuiz(page, 'settings');
    await root(page).locator('select[name="reveal.mode"]').selectOption('after_finish');
    await root(page).locator('select[name="hints.pricing"]').selectOption('free');
    await root(page).locator('input[name="limits.dailyPlays"]').fill('۵۰');
    await root(page).locator('[data-quiz-save="settings"]').click();
    await expect(page.getByText('ذخیره شد')).toBeVisible();
    const s = await qz.settings();
    expect(s.reveal.mode).toBe('after_finish');
    expect(s.hints.pricing).toBe('free');
    expect(s.limits.dailyPlays).toBe(50);
    await root(page).locator('input[name="rewards.dayZone"]').fill('Mars/Base');
    await root(page).locator('[data-quiz-save="settings"]').click();
    await expect(root(page).locator('[data-quiz-msg]')).toContainText('منطقهٔ زمانی شناخته نشد');

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['packs', 'Quiz packs'], ['questions', 'Questions'], ['missions', 'Missions and games'], ['reports', 'Player reports'], ['results', 'Result reports'],
      ['import', 'Import and export'], ['stats', 'Quiz report'], ['settings', 'Quiz settings']]) {
      await m.goto(CONSOLE_URL + '/#/quiz/' + sub);
      await expect(m.locator(`#quiz-root [data-quiz-page="${sub}"] h1`)).toHaveText(title);
      await noHorizontalScroll(m);
    }
  });
});
