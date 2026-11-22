import { expect, test } from '../../lib/fixtures';
import { eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { open, root, rows, search } from '../../lib/console';

const k = uid().replace(/_/g, '');
const prefix = 'e2ebot' + k.slice(-6);

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  sql(`DELETE FROM economy_game_stats WHERE user_id IN (SELECT u.id FROM users u WHERE u.username LIKE '${prefix}%')`);
  sql(`DELETE FROM economy_bot WHERE user_id IN (SELECT u.id FROM users u WHERE u.username LIKE '${prefix}%')`);
  sql(`DELETE FROM users WHERE username LIKE '${prefix}%'`);
});

test.describe('bots console', () => {
  test('bulk create, search, edit, disable and top up', async ({ page }) => {
    await open(page, '#/economy/bots');
    await expect(page.locator('a[href^="#/economy/bots"]').first()).toBeAttached();
    await expect(root(page).locator('[data-bot-kpis] [data-bot-kpi="total"]')).toBeVisible();
    await root(page).locator('[data-bot-action="bulk"]').click();
    const form = page.locator('[data-bot-bulk-form]');
    await form.locator('textarea').first().fill([prefix + 'a', prefix + 'b', prefix + 'c'].join('\n'));
    await page.locator('[data-bot-bulk-go]').click();
    await expect(page.locator('[data-bot-bulk-result]')).toContainText('۳ ربات ساخته شد');
    await page.keyboard.press('Escape');
    await search(page, prefix);
    await expect(rows(page)).toHaveCount(3);
    await rows(page).first().locator('[data-bot-row]').click();
    const edit = page.locator('[data-bot-edit]');
    await expect(edit.locator('[data-game-stats]')).toBeVisible();
    await page.locator('[data-bot-save]').click();
    await expect(page.locator('[data-bot-edit]')).toHaveCount(0);
    await expect(root(page).locator('[data-bot-bulkbar]')).toBeHidden();
    await root(page).locator('[data-select-all]').click();
    await expect(root(page).locator('[data-bot-bulkbar]')).toBeVisible();
    await expect(root(page).locator('[data-bot-selected]')).toContainText('۳');
    await root(page).locator('[data-bot-bulk="disable"]').click();
    await expect(rows(page).first()).toContainText('خاموش');
    const left = await eco.get('/bots?q=' + prefix + '&enabled=true');
    expect(left.total).toBe(0);
  });

  test('csv dry run shows problems per line, then import', async ({ page }) => {
    await open(page, '#/economy/bots');
    await root(page).locator('[data-bot-action="import"]').click();
    const f = page.locator('[data-bot-import-form]');
    await f.locator('textarea').first().fill(`username,level,template\n${prefix}x,7,hard\nbad name,3,\n`);
    await page.locator('[data-bot-import-dry]').click();
    await expect(page.locator('[data-bot-import-result] [data-bot-problem="INVALID_USERNAME"]')).toHaveCount(1);
    await page.locator('[data-bot-import-go]').click();
    await expect(page.locator('[data-bot-import-result]')).toContainText('۱ ربات ساخته شد');
    const got = await eco.get('/bots?q=' + prefix + 'x');
    expect(got.items[0].template).toBe('hard');
    expect(got.items[0].level).toBe(7);
  });

  test('game matches list opens and shows its filters', async ({ page }) => {
    await open(page, '#/economy/game-matches');
    await expect(root(page).locator('.eco-filters')).toBeVisible();
  });
});
