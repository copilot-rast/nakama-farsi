import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, rpc, uid, type Player } from '../../lib/api';
import { qz, removeQuiz } from '../../lib/quiz';
import { noHorizontalScroll } from '../../lib/console';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const CAT = 'e2e_' + k + '_cat';
const CAT2 = 'e2e_' + k + '_cat2';
const users: string[] = [];
const F: Record<string, any> = {};
let rawSettings = '';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#quiz-root');
const panel = (page: Page) => page.locator('[data-quiz=panel] [role=dialog]');

async function openQuiz(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/quiz/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-quiz-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

async function player(tag: string): Promise<Player> {
  const p = await devicePlayer(`e2e-qd-${tag}-${k}`);
  users.push(p.id);
  return p;
}

async function ok(p: Player, id: string, body: any) {
  const r = await rpc(p, id, body);
  expect(r.status, id + ' ' + JSON.stringify(r.data)).toBe(200);
  return r.data;
}

async function playTurn(p: Player, duelId: string, right: boolean) {
  for (let i = 0; i < 20; i++) {
    const v = await ok(p, 'quiz_duel_get', { duelId });
    if (v.step.kind === 'pick') {
      await ok(p, 'quiz_duel_pick_category', { duelId, categoryKey: v.step.offers[0].key });
      continue;
    }
    if (v.step.kind !== 'answer') return v;
    const q = (await ok(p, 'quiz_duel_question', { duelId })).question;
    await ok(p, 'quiz_duel_answer', { duelId, questionId: q.questionId, optionId: right ? F.right[q.questionId] : F.wrong[q.questionId] });
  }
  throw new Error('turn did not end');
}

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM quiz_settings WHERE id=1`).trim();
  F.hidden = sql(`UPDATE quiz_category SET duel=false WHERE duel AND key NOT LIKE 'e2e_${k}%' RETURNING key`).split('\n').map((x) => x.trim()).filter((x) => x && !/^UPDATE/.test(x));
  F.right = {};
  F.wrong = {};
});

test.afterAll(async () => {
  sql(`DELETE FROM quiz_duel WHERE player_a::text IN (${users.map((u) => `'${u}'`).join(',') || "''"}) OR player_b::text IN (${users.map((u) => `'${u}'`).join(',') || "''"});
    DELETE FROM quiz_duel_answer WHERE user_id::text IN (${users.map((u) => `'${u}'`).join(',') || "''"});
    DELETE FROM quiz_duel_day WHERE user_id::text IN (${users.map((u) => `'${u}'`).join(',') || "''"});
    DELETE FROM quiz_category WHERE key LIKE 'e2e_${k}%';
    UPDATE quiz_category SET duel=true WHERE key = ANY(ARRAY[${(F.hidden || []).map((x: string) => `'${x}'`).join(',') || "''"}]);`);
  for (const id of users) await deleteUser(id);
  removeQuiz('e2e_' + k);
  if (rawSettings) sql(`UPDATE quiz_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
});

test.describe('quiz duels console', () => {
  test('duel categories: create with a color, too few questions until three are added, IN_USE delete', async ({ page }) => {
    await openQuiz(page, 'duel-categories');
    await expect(page.locator('[data-quiz-group] a[href="#/quiz/duel-categories"]')).toBeAttached();
    await expect(page.locator('[data-pl-mode="quiz"]')).toBeVisible();
    await root(page).locator('[data-quiz-action="new-duel-category"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill(CAT);
    await p.locator('input[name="name-fa"]').fill('تاریخ ' + k);
    await p.locator('input[name="color"]').fill('red');
    await p.locator('[data-quiz-action="save-duel-category"]').click();
    await expect(p).toContainText('INVALID_COLOR');
    await p.locator('input[name="color"]').fill('#22AA66');
    await p.locator('[data-quiz-action="save-duel-category"]').click();
    await expect(p).toHaveCount(0);
    const row = root(page).locator(`tr[data-quiz-row="${CAT}"]`);
    await expect(row).toContainText('تاریخ ' + k);
    await expect(row).toContainText('کمتر از ۳ سوال');
    for (let i = 0; i < 3; i++) {
      const q = (await qz.post('/questions', { type: 'options', category: CAT, question: { fa: 'سوال دوئل ' + i + ' ' + k },
        options: [{ id: 0, text: { fa: 'نه' } }, { id: 1, text: { fa: 'آری' }, correct: true }] })).question;
      F.right[q.id] = 1;
      F.wrong[q.id] = 0;
    }
    await page.reload();
    await expect(root(page).locator(`tr[data-quiz-row="${CAT}"]`)).toContainText('قابل بازی');
    await expect(root(page).locator(`tr[data-quiz-row="${CAT}"]`)).toContainText('۳ سوال فعال');
    await root(page).locator(`[data-quiz-delete="${CAT}"]`).click();
    await page.locator('[data-quiz-confirm="ok"]').click();
    await expect(page.locator('[data-quiz=confirm]')).toContainText('بهتر است خاموشش کنید');
    await page.getByRole('button', { name: 'انصراف' }).last().click();
    const cats = await qz.get('/duel-categories?q=' + CAT);
    expect(cats.items.find((c: any) => c.key === CAT)).toMatchObject({ color: '#22AA66', questions: 3, playable: true, duel: true });
  });

  test('duel settings: Persian digits are saved, a bad value names its field', async ({ page }) => {
    await openQuiz(page, 'settings');
    const box = root(page).locator('[data-quiz-settings="duel"]');
    await box.locator('input[name="duel.rounds"]').fill('۲');
    await box.locator('input[name="duel.questionsPerRound"]').fill('۱');
    await box.locator('input[name="duel.categoryChoices"]').fill('۱');
    await box.locator('input[name="duel.turnHours"]').fill('۱۲');
    await box.locator('select[name="duel.timeoutResult"]').selectOption('score');
    await root(page).locator('[data-quiz-save="settings"]').click();
    await expect(page.getByText('ذخیره شد')).toBeVisible();
    const s = await qz.settings();
    expect(s.duel).toMatchObject({ rounds: 2, questionsPerRound: 1, categoryChoices: 1, turnHours: 12, timeoutResult: 'score', questionSeconds: 20, friendsOnly: true });
    await box.locator('input[name="duel.rounds"]').fill('۰');
    await root(page).locator('[data-quiz-save="settings"]').click();
    await expect(root(page).locator('[data-quiz-msg]')).toContainText('duel.rounds');
  });

  test('a real duel between two players shows in the list and the detail; an open duel is cancelled from the console', async ({ page }) => {
    const a = await player('a');
    const b = await player('b');
    const cats = await ok(a, 'quiz_duel_categories', {});
    expect(cats.rules).toMatchObject({ rounds: 2, questionsPerRound: 1 });
    expect(cats.categories.find((c: any) => c.key === CAT)).toMatchObject({ name: 'تاریخ ' + k, color: '#22AA66', playable: true });
    const start = await ok(a, 'quiz_duel_start', { mode: 'random' });
    F.duel = start.duelId;
    expect(start.step).toMatchObject({ kind: 'pick', round: 1 });
    await playTurn(a, F.duel, true);
    const joined = await ok(b, 'quiz_duel_start', { mode: 'random' });
    expect(joined).toMatchObject({ duelId: F.duel, joined: true, status: 'active', turn: 'you' });
    await playTurn(b, F.duel, false);
    const fin = await playTurn(a, F.duel, true);
    expect(fin).toMatchObject({ status: 'finished', result: 'win', reason: 'score' });

    const c = await player('c');
    F.open = (await ok(c, 'quiz_duel_start', { mode: 'random' })).duelId;

    await openQuiz(page, 'duels');
    await expect(root(page).locator('[data-quiz-kpi="started"]')).toContainText(/[۰-۹]/);
    await root(page).locator('input[name="player"]').fill(a.id);
    const row = root(page).locator(`tr[data-quiz-row="${F.duel}"]`);
    await expect(row).toContainText('تمام‌شده');
    await expect(row).toContainText('برنده');
    await row.click();
    const p = panel(page);
    await expect(p.locator('[data-quiz-section="round-1"]')).toContainText('تاریخ ' + k);
    await expect(p.locator('[data-quiz-section="round-1"]')).toContainText('درست');
    await expect(p.locator('[data-quiz-section="round-2"]')).toContainText('غلط');
    await expect(p.locator('[data-quiz-action="cancel-duel"]')).toBeDisabled();
    await page.keyboard.press('Escape');

    await root(page).locator('input[name="player"]').fill(c.id);
    await root(page).locator(`tr[data-quiz-row="${F.open}"]`).click();
    await panel(page).locator('[data-quiz-action="cancel-duel"]').click();
    await page.locator('[data-quiz-confirm="ok"]').click();
    await expect(page.getByText('دوئل لغو شد')).toBeVisible();
    await expect(root(page).locator(`tr[data-quiz-row="${F.open}"]`)).toContainText('لغوشده');
    const v = await ok(c, 'quiz_duel_get', { duelId: F.open });
    expect(v).toMatchObject({ status: 'cancelled', reason: 'admin' });
  });

  test('Ctrl K finds the duel pages; 390px English dark without sideways scroll', async ({ page, spawn }) => {
    await openQuiz(page, 'packs');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('دسته‌های دوئل');
    await expect(page.locator('.pl-pal-row').first()).toContainText('دسته‌های دوئل');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/quiz\/duel-categories/);
    await page.reload();
    await expect(root(page).locator('[data-quiz-page="duel-categories"]')).toBeVisible();
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('دوئل‌ها');
    await expect(page.locator('.pl-pal-row').first()).toContainText('دوئل‌ها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/quiz\/duels/);

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['duels', 'Duels'], ['duel-categories', 'Duel categories'], ['settings', 'Quiz settings']]) {
      await m.goto(CONSOLE_URL + '/#/quiz/' + sub);
      await expect(m.locator(`#quiz-root [data-quiz-page="${sub}"] h1`)).toHaveText(title);
      await expect(m.getByText('Loading…')).toHaveCount(0);
      await noHorizontalScroll(m);
    }
    const f = await spawn({ signedIn: true, mobile: true, width: 390, height: 844 });
    await f.goto(CONSOLE_URL + '/#/quiz/duels?player=' + users[0]);
    await expect(f.locator(`#quiz-root tr[data-quiz-row="${F.duel}"]`)).toBeVisible();
    await noHorizontalScroll(f);
    await f.locator(`#quiz-root tr[data-quiz-row="${F.duel}"]`).click();
    await expect(f.locator('[data-quiz=panel] [data-quiz-section="round-1"]')).toBeVisible();
    await noHorizontalScroll(f);
  });
});
