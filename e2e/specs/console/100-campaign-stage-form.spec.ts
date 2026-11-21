import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, eco, uid } from '../../lib/api';
import { type CampaignSnapshot, cg, plainCampaign, removeCampaign, restoreCampaignSettings, saveCampaignSettings } from '../../lib/campaign';
import { noHorizontalScroll } from '../../lib/console';
import { API_URL, CONSOLE_URL, ROOT, SERVER_KEY } from '../../lib/env';

const k = uid().replace(/_/g, '');
const F: Record<string, any> = {};
const users: string[] = [];
let saved: CampaignSnapshot = null;
const CH = 'e2e_' + k + '_police';
const ST = 'e2e_' + k + '_bomb';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#campaign-root');
const panel = (page: Page) => page.locator('[data-campaign=panel] [role=dialog]');

async function openCampaign(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/campaign/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-campaign-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

const stageFields = [
  { key: 'unityLevelId', label: { fa: 'شناسهٔ مرحله در یونیتی', en: 'Unity level id' }, type: 'text' },
  { key: 'brief', label: { fa: 'خلاصهٔ مأموریت', en: 'Mission brief' }, type: 'i18n' },
  { key: 'showArrowGuidance', label: { fa: 'فلش راهنما', en: 'Arrow guidance' }, type: 'bool', default: true },
  { key: 'missions', label: { fa: 'مأموریت', en: 'Mission' }, type: 'list', maxItems: 4, fields: [
    { key: 'mode', label: { fa: 'حالت', en: 'Mode' }, type: 'select', options: ['target', 'capture'], default: 'target' },
    { key: 'reachRadius', label: { fa: 'شعاع رسیدن', en: 'Reach radius' }, type: 'number', default: 8, min: 0, max: 500 },
    { key: 'enemies', label: { fa: 'دشمن', en: 'Enemy' }, type: 'list', fields: [{ key: 'speedMax', label: { fa: 'بیشترین سرعت', en: 'Top speed' }, type: 'number' }] },
  ] },
];

test.beforeAll(async () => {
  saved = saveCampaignSettings();
  F.cur = 'e2e' + k + 'c';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ فرم ' + k, en: 'Form coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  await cg.setSettings((s) => {
    plainCampaign(s);
    s.attempts.defaultMinSeconds = 0;
    s.fields = { stage: stageFields, chapter: [{ key: 'offerEvery', label: { fa: 'فاصلهٔ پیشنهاد', en: 'Offer every' }, type: 'integer', default: 120, chapters: [CH] }] };
  });
  F.ch = await cg.post('/chapters', { key: CH, name: { fa: 'کلانتری ' + k }, isActive: true, sort: 1000,
    milestones: [{ percent: 16.6667, currencies: { [F.cur]: 20 } }, { percent: 100, currencies: { [F.cur]: 80 } }] });
});

test.afterAll(async () => {
  for (const id of users) await deleteUser(id);
  removeCampaign('e2e_' + k);
  await restoreCampaignSettings(saved);
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('campaign stage form', () => {
  test('chapter steps keep ids; the chapter settings form saves its field', async ({ page }) => {
    expect(F.ch.milestones.map((m: any) => m.id)).toEqual(['1', '2']);
    await openCampaign(page, 'chapters');
    await root(page).locator(`tr[data-campaign-row="${CH}"]`).click();
    const p = panel(page);
    await expect(p.locator('[data-campaign-milestone-id="1"]')).toHaveText('شناسهٔ پله: ۱');
    await p.locator('[data-campaign-milestone]').first().locator('input[name="milestone-percent"]').fill('20');
    await p.locator('input[name="settings.offerEvery"]').fill('۹۰');
    await p.locator('[data-campaign-action="save-chapter"]').click();
    await expect(p).toHaveCount(0);
    const ch = await cg.get('/chapters/' + F.ch.id);
    expect(ch.milestones[0]).toMatchObject({ id: '1', percent: 20 });
    expect(ch.settings).toEqual({ offerEvery: 90 });
  });

  test('a stage from the panel with a template, no replay and the mission form', async ({ page }) => {
    await openCampaign(page, 'stages?chapterId=' + F.ch.id);
    await root(page).locator('[data-campaign-action="new-stage"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill(ST);
    await p.locator('input[name="name-fa"]').fill('بمب‌گذاری ' + k);
    await p.locator('input[name="timeLimitSeconds"]').fill('۶۰');
    await p.locator('input[name="minSeconds"]').fill('۳');
    await p.locator('input[name="noReplayAfterWin"]').check();
    await p.locator('select[name="starTemplate"]').selectOption('race');
    await expect(p.locator('[data-campaign-template-rules] li').first()).toContainText('۳ ستاره وقتی');
    await expect(p.locator('[data-campaign-template-rules] li').first()).not.toContainText('سلامت');
    await p.locator('input[name="settings.unityLevelId"]').fill('M_Bombing_' + k);
    await p.locator('textarea[name="settings.brief-fa"]').fill('عملیات شبیخون');
    await expect(p.locator('input[name="settings.showArrowGuidance"]')).toBeChecked();
    await p.locator('[data-campaign-action="add-missions"]').click();
    const mission = p.locator('[data-campaign-list-item="missions"]').first();
    await expect(mission.locator('[data-campaign-list-title]')).toHaveText('مأموریت ۱');
    await mission.locator('select[name="missions.mode"]').selectOption('capture');
    await mission.locator('input[name="missions.reachRadius"]').fill('۱۲');
    await mission.locator('[data-campaign-action="add-enemies"]').click();
    await mission.locator('[data-campaign-list-item="enemies"] input[name="enemies.speedMax"]').fill('170');
    await p.locator('[data-campaign-currencies="stage"] + button').click();
    await p.locator('[data-campaign-currencies="stage"] select[name="reward-currency"]').selectOption(F.cur);
    await p.locator('[data-campaign-currencies="stage"] input[name="reward-amount"]').fill('1000');
    await p.locator('input[name="xp"]').fill('100');
    await mission.locator('input[name="missions.reachRadius"]').fill('۹۰۰');
    await p.locator('[data-campaign-action="save-stage"]').click();
    await expect(p.locator('.eco-notice')).toContainText('مقدار بیرون از بازهٔ مجاز است');
    await mission.locator('input[name="missions.reachRadius"]').fill('۱۲');
    await p.locator('[data-campaign-action="save-stage"]').click();
    await expect(p).toHaveCount(0);
    const st = (await cg.get('/stages?q=' + ST)).items[0];
    expect(st).toMatchObject({ starTemplate: 'race', noReplayAfterWin: true, minSeconds: 3, timeLimitSeconds: 60, ownStarRules: false });
    expect(st.settings).toEqual({ unityLevelId: 'M_Bombing_' + k, brief: { fa: 'عملیات شبیخون' }, missions: [{ mode: 'capture', reachRadius: 12, enemies: [{ speedMax: 170 }] }] });
    F.st = st;

    await root(page).locator(`tr[data-campaign-row="${ST}"]`).click();
    await expect(p.locator('select[name="starTemplate"]')).toHaveValue('race');
    await expect(p.locator('[data-campaign-list-item="missions"] input[name="missions.reachRadius"]')).toHaveValue('۱۲');
    await p.getByRole('button', { name: 'انصراف' }).click();
  });

  test('a live nakama-js player reads the form data, plays and claims the stage, then cannot replay it', async () => {
    const out = execFileSync(process.execPath, [path.join(ROOT, 'live', 'campaign-play.mjs')], {
      cwd: ROOT, env: { ...process.env, API_URL, SERVER_KEY, STAGE_KEY: ST }, encoding: 'utf8', timeout: 60_000,
    });
    const run = JSON.parse(out.trim().split('\n').pop()!);
    users.push(run.userId);
    expect(run.stage.settings).toEqual({ unityLevelId: 'M_Bombing_' + k, brief: 'عملیات شبیخون', showArrowGuidance: true,
      missions: [{ mode: 'capture', reachRadius: 12, enemies: [{ speedMax: 170 }] }] });
    expect(run.stage).toMatchObject({ starTemplate: 'race', minSeconds: 3, noReplayAfterWin: true, replayable: true });
    expect(run.stage.chapterSettings).toEqual({ offerEvery: 90 });
    expect(run.stage.starRules.flatMap((r: any) => r.conditions.map((c: any) => c.field))).not.toContain('healthPercent');
    expect(run.result).toMatchObject({ stars: 2, rewards: [{ currencyKey: F.cur, amount: 500 }], xpGained: 50, minSeconds: 3, replayable: false, milestoneIds: ['1'] });
    const tiers = Object.fromEntries(run.result.criteria.map((c: any) => [c.field, c.stars]));
    expect(tiers).toEqual({ timeSavedPercent: 3, timesSeen: 2 });
    expect(run.replay).toBe('STAGE_ALREADY_WON');
  });

  test('settings page edits templates and the free ride table switch; the form fits 390px', async ({ page, spawn }) => {
    await openCampaign(page, 'settings');
    await expect(root(page).locator('[data-campaign-template]')).toHaveCount(4);
    await expect(root(page).locator('input[name="freeride.tableEnabled"]')).not.toBeChecked();
    await expect(page.locator('[data-campaign-group] a[href="#/campaign/freeride"]')).toHaveCount(0);
    await root(page).locator('[data-campaign-template="quest"] input[name="template-name-fa"]').fill('کوئست ' + k);
    await root(page).locator('[data-campaign-save="settings"]').click();
    await expect(page.getByText('ذخیره شد')).toBeVisible();
    expect((await cg.settings()).stars.templates.find((x: any) => x.key === 'quest').name.fa).toBe('کوئست ' + k);

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await m.goto(CONSOLE_URL + '/#/campaign/stages?chapterId=' + F.ch.id);
    await m.locator(`#campaign-root tr[data-campaign-row="${ST}"]`).click();
    await expect(m.locator('[data-campaign=panel] [data-campaign-section="settings"] h2')).toHaveText('Mission settings');
    await noHorizontalScroll(m);
  });
});
