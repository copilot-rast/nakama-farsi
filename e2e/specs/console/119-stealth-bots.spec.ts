import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { base, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, open, root, rows, search, toasts } from '../../lib/console';

const k = uid().replace(/[^a-z0-9]/g, '').slice(-8);
const key = 'e2e_sb_' + k;
const prefix = 'e2esb' + k.slice(-6);

test.describe.configure({ mode: 'serial' });

async function toggleOn(scope: Locator, label: string) {
  const box = scope.locator('.eco-field', { hasText: label }).first();
  await box.getByText(label, { exact: true }).click();
  await expect(box.locator('input[role="switch"]')).toHaveAttribute('aria-checked', 'true');
}

test.beforeAll(async () => {
  await eco.post('/games', { key, isActive: true, name: { fa: 'بازی آزمون بات', en: 'Bot test game' }, minPlayers: 2, maxPlayers: 2, botsAllowed: true, rooms: [] });
  await eco.post('/bots/bulk', { names: [prefix + 'a', prefix + 'b'], levelMin: 3, levelMax: 3, template: 'normal', wallet: {} });
});

test.afterAll(async () => {
  sql(`DELETE FROM economy_bot_lease WHERE user_id IN (SELECT id FROM users WHERE username LIKE '${prefix}%')`);
  sql(`DELETE FROM economy_bot WHERE user_id IN (SELECT id FROM users WHERE username LIKE '${prefix}%')`);
  sql(`DELETE FROM users WHERE username LIKE '${prefix}%'`);
  sql(`DELETE FROM economy_game WHERE key='${key}'`);
});

test.describe('stealth bots console', () => {
  test('turn stealth on for one game, set the timing and see it saved', async ({ page }) => {
    await open(page, '#/economy/bots/stealth?game=' + key);
    const r = root(page);
    await expect(r.locator('[data-settings-form="stealth-bots"]')).toBeVisible();
    await expect(r.locator('[data-sb-summary]')).toContainText('خاموش');
    await toggleOn(r, 'بات پنهان روشن باشد');
    await expect(field(r, 'کمترین انتظار تا آمدن حریف (ثانیه)')).toBeVisible();
    await field(r, 'کمترین انتظار تا آمدن حریف (ثانیه)').locator('input').fill('20');
    await field(r, 'بیشترین انتظار (ثانیه)').locator('input').fill('5');
    await r.locator('[data-settings-save]').click();
    await expect(field(r, 'بیشترین انتظار (ثانیه)')).toContainText('بیشترین مقدار از کمترین کمتر است');
    await field(r, 'بیشترین انتظار (ثانیه)').locator('input').fill('25');
    await toggleOn(r, 'محافظ روزانه روشن باشد');
    await field(r, 'برد روزانه پیش از قوی‌تر شدن').locator('input').fill('8');
    await r.locator('[data-settings-save]').click();
    await expect(toasts(page).first()).toBeVisible();
    await expect(r.locator('[data-sb-summary]')).toContainText('۲۰ تا ۲۵');
    await expect(r.locator('[data-sb-pool] [data-sb-stat="pool"]')).toBeVisible();
    const saved = await eco.get('/game-stealth/' + key);
    expect(saved.settings.enabled).toBe(true);
    expect(saved.settings.arriveMinSeconds).toBe(20);
    expect(saved.settings.arriveMaxSeconds).toBe(25);
    expect(saved.settings.guard).toBe(true);
    expect(saved.settings.guardDailyWins).toBe(8);
    expect(saved.settings.trophies).toBe(true);
  });

  test('the game form links to the section and the bots page has the tab and status', async ({ page }) => {
    await open(page, '#/economy/bots/stealth?game=' + key);
    await expect(page.locator('a[href="#/economy/bots"]').first()).toBeAttached();
    await open(page, '#/economy/games');
    await search(page, key);
    await rows(page).first().click();
    await expect(page.locator('[data-gm-stealth="on"] a[href^="#/economy/bots/stealth?game=' + key + '"]')).toBeVisible();
    await open(page, '#/economy/bots');
    await search(page, prefix);
    await expect(rows(page)).toHaveCount(2);
    await expect(rows(page).first().locator('[data-bot-now="idle"]')).toBeVisible();
    sql(`INSERT INTO economy_bot_lease (user_id, match_id, game_key, busy) SELECT id, 'm-${k}', '${key}', true FROM users WHERE username='${prefix}a' ON CONFLICT (user_id) DO UPDATE SET busy=true, leased_at=now()`);
    await page.reload();
    await search(page, prefix + 'a');
    await expect(rows(page).first().locator('[data-bot-now="match"]')).toBeVisible();
  });

  test('ctrl+k finds the section', async ({ page }) => {
    await open(page, '#/economy/bots');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('بات پنهان');
    await expect(page.locator('.pl-pal-row[href^="#/economy/bots/stealth"]').first()).toBeVisible();
  });

  test('pool growth, level look and one match are settings with defaults', async ({ page }) => {
    await open(page, '#/economy/bots/stealth?game=' + key);
    const r = root(page);
    await expect(r.locator('[data-settings-form="stealth-bots"]')).toBeVisible();
    await expect(field(r, 'بیشترین تعداد ربات‌های استخر').locator('input')).toHaveValue(/2[,٬]?000|۲[,٬]?۰۰۰/);
    await expect(field(r, 'گوناگونی ظاهر (درصد)')).toBeVisible();
    await expect(r.locator('[data-sb-summary]')).toContainText('خودش بزرگ می‌شود');
    await field(r, 'بیشترین تعداد ربات‌های استخر').locator('input').fill('150');
    await field(r, 'گوناگونی ظاهر (درصد)').locator('input').fill('5');
    await r.locator('[data-settings-save]').click();
    await expect(toasts(page).first()).toBeVisible();
    const saved = await eco.get('/game-stealth/' + key);
    expect(saved.settings.poolMax).toBe(150);
    expect(saved.settings.lookSpread).toBe(5);
    expect(saved.settings.autoGrow).toBe(true);
    expect(saved.settings.lookByLevel).toBe(true);
    expect(saved.settings.oneLiveMatch).toBe(true);
    expect(saved.settings.presences).toBe(true);
    expect(saved.defaults.poolMax).toBe(2000);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('ربات تازه بساز');
    await expect(page.locator('.pl-pal-row[href^="#/economy/bots/stealth"]').first()).toBeVisible();
  });

  test('friend rules hold the random decline range', async ({ page }) => {
    const before = await consoleRules();
    await open(page, '#/project-settings/friends');
    const r = root(page);
    await field(r, 'رد درخواست به ربات، کمترین انتظار (دقیقه)').locator('input').fill('30');
    await field(r, 'بیشترین انتظار (دقیقه)').locator('input').fill('240');
    await r.locator('[data-settings-form]', { hasText: 'رد درخواست به ربات، کمترین انتظار' }).locator('[data-settings-save]').click();
    await expect(toasts(page).first()).toBeVisible();
    const after = await consoleRules();
    expect(after.botDeclineMinMinutes).toBe(30);
    expect(after.botDeclineMaxMinutes).toBe(240);
    sql(`UPDATE base_friend_rules SET bot_decline_min_minutes=${before.botDeclineMinMinutes || 0}, bot_decline_max_minutes=${before.botDeclineMaxMinutes || 0} WHERE id=1`);
  });
});

async function consoleRules() {
  return base.get('/friend-rules');
}
