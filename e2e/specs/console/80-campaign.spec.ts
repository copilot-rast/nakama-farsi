import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { type CampaignSnapshot, cg, plainCampaign, removeCampaign, restoreCampaignSettings, saveCampaignSettings } from '../../lib/campaign';
import { noHorizontalScroll } from '../../lib/console';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const F: Record<string, any> = {};
const users: string[] = [];
let saved: CampaignSnapshot = null;
const CH = 'e2e_' + k + '_story';
const ST = 'e2e_' + k + '_chase';
const ST2 = 'e2e_' + k + '_second';
const FR = 'e2e_' + k + '_roam';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#campaign-root');
const panel = (page: Page) => page.locator('[data-campaign=panel] [role=dialog]');

async function openCampaign(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/campaign/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-campaign-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

test.beforeAll(async () => {
  saved = saveCampaignSettings();
  F.cur = 'e2e' + k + 'c';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ کنسول ' + k, en: 'Console coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  F.prize = await eco.post('/prizes', { key: 'e2e_' + k + '_medal', name: { fa: 'مدال ' + k }, message: {}, tagIds: [], isRepeatable: true, metadata: {}, items: [{ type: 'currency', key: F.cur, action: 'add', value: 1 }] });
  await cg.setSettings((s) => { plainCampaign(s); s.attempts.defaultMinSeconds = 0; });
});

test.afterAll(async () => {
  for (const id of users) await deleteUser(id);
  removeCampaign('e2e_' + k);
  await restoreCampaignSettings(saved);
  if (F.prize) { sql(`DELETE FROM economy_user_prize WHERE prize_id='${F.prize.id}'`); await eco.del('/prizes/' + F.prize.id); }
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('campaign console', () => {
  test('sidebar group; a chapter with an unlock rule and a star milestone from the editor', async ({ page }) => {
    await openCampaign(page, 'chapters');
    await expect(page.locator('[data-campaign-group] a[href="#/campaign/stages"]')).toBeAttached();
    await root(page).locator('[data-campaign-action="new-chapter"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill(CH);
    await p.locator('input[name="name-fa"]').fill('فصل داستان ' + k);
    await p.locator('input[name="address"]').fill('Chapters/Story');
    await p.locator('[data-campaign-rules="chapter"] + button, [data-campaign-action="add-rule"]').first().click();
    const rule = p.locator('[data-campaign-rules="chapter"] .cg-row').first();
    await rule.locator('select[name="rule-type"]').selectOption('level');
    await rule.locator('input[name="rule-value"]').fill('۲');
    await expect(rule.locator('[data-campaign-rule-sentence]')).toHaveText('سطح ≥ ۲');
    await p.locator('[data-campaign-action="add-milestone"]').click();
    const ms = p.locator('[data-campaign-milestone]').first();
    await ms.locator('input[name="milestone-percent"]').fill('50');
    await ms.locator('[data-campaign-action="add-currency"]').click();
    await ms.locator('select[name="reward-currency"]').selectOption(F.cur);
    await ms.locator('input[name="reward-amount"]').fill('40');
    await p.locator('[data-campaign-action="save-chapter"]').click();
    await expect(p).toHaveCount(0);
    const row = root(page).locator(`tr[data-campaign-row="${CH}"]`);
    await expect(row).toContainText('فصل داستان');
    await expect(row).toContainText('سطح ≥ ۲');
    const ch = (await cg.get('/chapters')).items.find((c: any) => c.key === CH);
    expect(ch.unlock).toEqual([{ type: 'level', value: 2 }]);
    expect(ch.milestones).toEqual([{ percent: 50, currencies: { [F.cur]: 40 } }]);
    expect(ch.address).toBe('Chapters/Story');
    F.ch = ch;
    await cg.put('/chapters/' + ch.id, { ...ch, unlock: [] });
  });

  test('a stage with its own star rule, a payout preview and a star prize; a key error shows on the field', async ({ page }) => {
    await openCampaign(page, 'stages?chapterId=' + F.ch.id);
    await root(page).locator('[data-campaign-action="new-stage"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill('bad key!');
    await p.locator('input[name="name-fa"]').fill('مرحلهٔ ویژه ' + k);
    await p.locator('[data-campaign-action="save-stage"]').click();
    await expect(p.locator('.eco-notice')).toContainText('کلید درست نیست');
    await p.locator('input[name="key"]').fill(ST);
    await expect(p.locator('select[name="chapterId"]')).toHaveValue(F.ch.id);
    await p.locator('input[name="timeLimitSeconds"]').fill('۶۰');
    await p.locator('[data-campaign-currencies="stage"] + button').click();
    await p.locator('[data-campaign-currencies="stage"] select[name="reward-currency"]').selectOption(F.cur);
    await p.locator('[data-campaign-currencies="stage"] input[name="reward-amount"]').fill('1000');
    await p.locator('input[name="xp"]').fill('100');
    await expect(p.locator('[data-campaign-preview] li').nth(1)).toContainText('۲ ستاره:');
    await expect(p.locator('[data-campaign-preview] li').nth(1)).toContainText(`۵۰۰ ${F.cur} + ۵۰ XP`);
    await p.locator('input[name="useDefaultStars"]').uncheck();
    const rules = p.locator('[data-campaign-star-rule]');
    await expect(rules).toHaveCount(2);
    await rules.nth(1).getByRole('button', { name: 'حذف' }).first().click();
    const first = rules.first();
    await first.locator('select[name="cond-field"]').first().selectOption('__custom');
    await first.locator('input[name="cond-custom"]').first().fill('arrests');
    await first.locator('input[name="cond-value"]').first().fill('5');
    await expect(first.locator('[data-campaign-star-sentence]')).toContainText('۳ ستاره وقتی arrests >= ۵');
    await p.locator('[data-campaign-action="add-tier"]').click();
    await p.locator('[data-campaign-tier] select[name="reward-prize"]').selectOption(F.prize.key);
    await p.locator('[data-campaign-action="save-stage"]').click();
    await expect(p).toHaveCount(0);
    await expect(root(page).locator(`tr[data-campaign-row="${ST}"] [data-campaign-reward="${ST}"]`)).toContainText('XP');
    const st = (await cg.get('/stages?q=' + ST)).items[0];
    expect(st).toMatchObject({ mode: 'ordered', timeLimitSeconds: 60, xp: 100, ownStarRules: true, ownUnlock: false, rewards: { currencies: { [F.cur]: 1000 } } });
    expect(st.starRules[0].conditions[0]).toEqual({ field: 'arrests', op: '>=', value: 5 });
    expect(st.tierPrizes).toEqual([{ stars: 3, prizeKey: F.prize.key }]);
    F.st = st;
    F.st2 = await cg.post('/stages', { key: ST2, chapterId: F.ch.id, name: { fa: 'دوم ' + k }, isActive: true, sort: 2000 });
  });

  test('a free ride reward with a cooldown and a daily cap', async ({ page }) => {
    await openCampaign(page, 'freeride');
    await root(page).locator('[data-campaign-action="new-freeride"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill(FR);
    await p.locator('input[name="name-fa"]').fill('گردش ' + k);
    await p.locator('[data-campaign-currencies="freeride"] + button').click();
    await p.locator('[data-campaign-currencies="freeride"] select[name="reward-currency"]').selectOption(F.cur);
    await p.locator('[data-campaign-currencies="freeride"] input[name="reward-amount"]').fill('5');
    await p.locator('input[name="cooldownSeconds"]').fill('600');
    await p.locator('input[name="dailyLimit"]').fill('3');
    await p.locator('[data-campaign-action="save-freeride"]').click();
    await expect(p).toHaveCount(0);
    await expect(root(page).locator(`[data-campaign-limits="${FR}"]`)).toHaveText('هر ۶۰۰ ثانیه · ۳ بار در روز');
  });

  test('report and player card after a player plays; a stage is reset from the card', async ({ page }) => {
    const pl: Player = await devicePlayer('e2e-cgc-' + k);
    users.push(pl.id);
    const a = (await rpc(pl, 'campaign_attempt_start', { stageKey: ST })).data;
    sql(`UPDATE campaign_attempt SET started_at = started_at - interval '40 seconds' WHERE id='${a.attemptId}'`);
    const done = await rpc(pl, 'campaign_attempt_complete', { attemptId: a.attemptId, success: true, elapsedSeconds: 30, fields: { arrests: 6 } });
    expect(done.data).toMatchObject({ stars: 3, rewards: [{ currencyKey: F.cur, amount: 1000 }] });
    const lost = (await rpc(pl, 'campaign_attempt_start', { stageKey: ST2 })).data;
    await rpc(pl, 'campaign_attempt_complete', { attemptId: lost.attemptId, success: false, elapsedSeconds: 1 });

    await openCampaign(page, 'report');
    await root(page).locator('select[name="chapterId"]').selectOption(F.ch.id);
    await expect(root(page).locator(`[data-campaign-players="${ST}"]`)).toHaveText('۱');
    await expect(root(page).locator(`[data-campaign-rate="${ST}"]`)).toHaveText('۱۰۰٪');
    await expect(root(page).locator(`[data-campaign-avg="${ST}"]`)).toHaveText('۳');
    await expect(root(page).locator(`[data-campaign-drop="${ST2}"]`)).toHaveText('۱۰۰٪');

    await page.goto(CONSOLE_URL + '/#/players/' + pl.id);
    const card = page.locator('[data-campaign-player]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    const row = card.locator(`[data-campaign-player-stage="${ST}"]`);
    await expect(row.locator('[data-campaign-stars="3"]')).toBeVisible();
    await row.locator(`[data-campaign-reset="${ST}"]`).click();
    await page.locator('[data-campaign-confirm="ok"]').click();
    await expect(card.locator(`[data-campaign-player-stage="${ST}"] [data-campaign-stars="0"]`)).toBeVisible();
    expect(sql(`SELECT best_stars || ':' || (claimed->>'${F.cur}') FROM campaign_user_stage WHERE user_id='${pl.id}' AND stage_id='${F.st.id}'`).trim()).toBe('0:1000');
  });

  test('settings: replay policy and XP meta are saved; pages fit 390px in English dark mode', async ({ page, spawn }) => {
    await openCampaign(page, 'settings');
    await root(page).locator('select[name="rewards.replayMode"]').selectOption('none');
    await root(page).locator('input[name="attempts.graceSeconds"]').fill('۵');
    await root(page).locator('[data-campaign-save="settings"]').click();
    await expect(page.getByText('ذخیره شد')).toBeVisible();
    const s = await cg.settings();
    expect(s.rewards.replayMode).toBe('none');
    expect(s.attempts.graceSeconds).toBe(5);
    await root(page).locator('input[name="rewards.dayZone"]').fill('Mars/Base');
    await root(page).locator('[data-campaign-save="settings"]').click();
    await expect(root(page).locator('[data-campaign-msg]')).toContainText('منطقهٔ زمانی شناخته نشد');

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['chapters', 'Chapters'], ['stages', 'Stages'], ['freeride', 'Free ride rewards'], ['report', 'Stage report'], ['settings', 'Stage settings']]) {
      await m.goto(CONSOLE_URL + '/#/campaign/' + sub);
      await expect(m.locator(`#campaign-root [data-campaign-page="${sub}"] h1`)).toHaveText(title);
      await noHorizontalScroll(m);
    }
  });
});
