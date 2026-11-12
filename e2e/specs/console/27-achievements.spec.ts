import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const key = 'e2e_' + k + '_ach';
const made = { users: [] as string[], prizes: [] as string[] };
let P: Player;
let prizeId = '';
const prizeName = 'سکهٔ دستاورد ' + k;

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

async function pickIn(scope: Locator, text: string) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

async function confirm(page: Page, label: string) {
  await page.locator('[role="alertdialog"] button', { hasText: label }).click();
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-ac-' + k, ('ac' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
  const p = await eco.post('/prizes', { key: 'e2e_' + k + '_coins', name: { fa: prizeName, en: 'Achievement coins ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 7 }] });
  prizeId = p.id;
  made.prizes.push(p.id);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_user_achievement WHERE achievement_id IN (SELECT id FROM economy_achievement WHERE key LIKE 'e2e\\_${k}%') OR user_id IN (${made.users.map((u) => `'${u}'`).join(',')})`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (SELECT id FROM economy_prize WHERE key LIKE 'e2e\\_${k}%')`);
  const all = await eco.get('/achievements?q=e2e_' + k + '&limit=100');
  const extra = await eco.get('/prizes?q=e2e_' + k + '&limit=100');
  for (const p of extra.items || []) if (!made.prizes.includes(p.id)) expect((await eco.del('/prizes/' + p.id)).status).toBe(200);
  for (const it of all.items || []) expect((await eco.del('/achievements/' + it.id)).status).toBe(200);
  for (const id of made.prizes) expect((await eco.del('/prizes/' + id)).status).toBe(200);
});

test.describe('achievements: console', () => {
  test('editor: event condition with a live sentence, tiers with prizes and order check, list row and locked key', async ({ page }) => {
    await open(page, '#/economy/achievements');
    await expect(root(page)).toContainText('نشان‌هایی که بازیکن');
    await expect(rows(page).filter({ hasText: 'ستاره‌چین پرکار' })).toHaveCount(1);
    await go(page, '#/economy/achievements/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('دستاورد جدید');
    await field(p, 'کلید').locator('input').fill(key);
    await fillTranslated(p, 'name', { fa: 'دونده ' + k, en: 'Runner ' + k });
    await fillTranslated(p, 'hint', { fa: 'چند دور بازی کن' });
    await field(p, 'کمیابی').locator('select').selectOption('epic');
    await field(p, 'امتیاز').locator('input').fill('25');
    await pickIn(field(p, 'رویداد'), 'پایان یک دور');
    const preview = p.locator('[data-ach-preview]');
    await expect(preview).toContainText('«پایان یک دور» ۱ بار');
    const tiers = p.locator('[data-rows="tiers"] [data-row]');
    await expect(tiers).toHaveCount(1);
    await tiers.nth(0).locator('[data-col="threshold"] input').fill('2');
    await pickIn(tiers.nth(0).locator('[data-col="prizeId"]'), prizeName);
    await p.locator('[data-rows-add="tiers"]').click();
    await expect(tiers).toHaveCount(2);
    await tiers.nth(1).locator('[data-col="threshold"] input').fill('1');
    await savePanel(page).catch(() => null);
    await expect(tiers.nth(1)).toContainText('هدف بزرگ‌تری');
    await tiers.nth(1).locator('[data-col="threshold"] input').fill('5');
    await pickIn(tiers.nth(1).locator('[data-col="prizeId"]'), prizeName);
    await tiers.nth(1).locator('[data-col="grantMode"] select').selectOption('auto');
    await expect(preview).toContainText('۲');
    await expect(preview).toContainText('۵');
    await savePanel(page);
    const a = await byKey('achievements', key);
    expect(a).toMatchObject({ source: 'event', metric: 'count', rarity: 'epic', points: 25, isHidden: false, hint: { fa: 'چند دور بازی کن' } });
    expect(a.tiers.map((x: any) => [x.tier, Number(x.threshold), x.prizeId, x.grantMode])).toEqual([[1, 2, prizeId, 'claim'], [2, 5, prizeId, 'auto']]);

    await search(page, key);
    await expect(rows(page)).toHaveCount(1);
    const row = rows(page).first();
    await expect(row).toContainText('دونده ' + k);
    await expect(row).toContainText('حماسی');
    await expect(row).toContainText('۲۵ امتیاز');
    await expect(row).toContainText('۲ رده');
    await expect(row.locator('[data-ach-sentence]')).toContainText('«پایان یک دور» ۲ بار');
    await row.locator('td').first().click();
    await settlePanel(page);
    await expect(panel(page)).toContainText('ویرایش دستاورد');
    await expect(field(panel(page), 'کلید').locator('input')).toHaveJSProperty('readOnly', true);
    await expect(panel(page).locator('[data-rows="tiers"] [data-row]')).toHaveCount(2);
    await page.keyboard.press('Escape');

    await go(page, '#/economy/achievements/new');
    await settlePanel(page);
    const q = panel(page);
    await field(q, 'کلید').locator('input').fill(key + '_task');
    await fillTranslated(q, 'name', { fa: 'زنجیره‌باز ' + k });
    await field(q, 'راه باز شدن').locator('select').selectOption('task');
    await expect(q.locator('[data-rows="tiers"]')).toBeHidden();
    await savePanel(page).catch(() => null);
    await expect(field(q, 'تسک')).toContainText('تسک یا گروه تسکی');
    await pickIn(field(q, 'یا گروه تسک'), 'زنجیرهٔ شکارچی');
    await q.locator('.x-toggle', { hasText: 'پنهان تا باز شدن' }).click();
    await savePanel(page);
    const b = await byKey('achievements', key + '_task');
    expect(b).toMatchObject({ source: 'task', isHidden: true, taskId: null });
    expect(b.taskGroupId).toBeTruthy();
    expect(b.tiers).toHaveLength(1);

    const tooFew = await consoleApi('PUT', '/v2/console/economy/achievements/' + a.id, Object.assign({}, a, { tiers: [], categoryId: '', taskId: '', taskGroupId: '' }));
    expect(tooFew.status).toBe(400);
    expect(tooFew.data.error).toBe('TIERS_REQUIRED');
  });

  test('prize editor gives an achievement, and that prize unlocks it for a player', async ({ page }) => {
    await open(page, '#/economy/prizes/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_badge');
    await fillTranslated(p, 'name', { fa: 'جعبهٔ نشان ' + k });
    const row = p.locator('.eco-items-row').first();
    await row.locator('select').first().selectOption('achievement');
    await pickIn(row.locator('[data-achievement-picker]'), 'زنجیره‌باز ' + k);
    await expect(row).toContainText('دستاورد برای بازیکن باز می‌شود');
    await savePanel(page);
    const prize = await byKey('prizes', 'e2e_' + k + '_badge');
    expect(prize.items).toEqual([{ type: 'achievement', key: key + '_task', action: 'give', value: 1 }]);
    await search(page, 'e2e_' + k + '_badge');
    await expect(rows(page).first()).toContainText('دستاورد: ' + key + '_task');
    await eco.post('/users/' + P.id + '/grant', { prizeId: prize.id });
    const list = await rpc(P, 'economy_achievements', { lang: 'fa' });
    expect(list.data.items.find((x: any) => x.key === key + '_task')).toMatchObject({ unlocked: true, state: 'completed' });
    const used = await consoleApi('DELETE', '/v2/console/economy/achievements/' + (await byKey('achievements', key + '_task')).id);
    expect(used.status).toBe(409);
    expect(used.data.error).toBe('IN_USE');
  });

  test('player card: progress from events, give the prize, give and take back an achievement, history', async ({ page }) => {
    await eco.post('/users/' + P.id + '/events', { key: 'round_finished', value: 10 });
    await eco.post('/users/' + P.id + '/events', { key: 'round_finished', value: 10 });
    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-achievements]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('h2')).toHaveText('دستاوردهای بازیکن');
    const mine = card.locator(`[data-pac="${key}"]`);
    await expect(mine).toHaveAttribute('data-state', 'unlocked');
    await expect(mine.locator('[data-pac-tier]')).toHaveText('ردهٔ ۱ از ۲');
    await expect(mine.locator('[data-pac-progress]')).toHaveText('۲ از ۵');
    await expect(mine.locator('[data-pac-by]')).toContainText('با بازی');
    await expect(card.locator(`[data-pac="${key}_task"] [data-pac-by]`)).toContainText('با جایزه');
    const before = (await rpc(P, 'economy_balances', {})).data.wallet.coin;
    await mine.locator(`[data-pac-claim="${key}"]`).click();
    await expect(toasts(page).last()).toContainText('جایزهٔ رده داده شد');
    await expect(card.locator(`[data-pac-claim="${key}"]`)).toHaveCount(0);
    expect((await rpc(P, 'economy_balances', {})).data.wallet.coin).toBe(before + 7);
    const again = await rpc(P, 'economy_achievement_claim', { key });
    expect(JSON.stringify(again.data)).toContain('ALREADY_CLAIMED');

    await pickIn(card.locator('[data-pac-give-row]'), 'هوادار ستاره‌چین');
    await card.locator('[data-pac-give]').click();
    await expect(toasts(page).last()).toContainText('دستاورد به بازیکن داده شد');
    const fan = card.locator('[data-pac="star_ach_supporter"]');
    await expect(fan).toHaveAttribute('data-state', 'completed');
    await expect(fan.locator('[data-pac-by]')).toContainText('از کنسول');
    await fan.locator('[data-pac-revoke="star_ach_supporter"]').click();
    await confirm(page, 'پس گرفتن');
    await expect(toasts(page).last()).toContainText('دستاورد پس گرفته شد');
    await expect(card.locator('[data-pac="star_ach_supporter"]')).toHaveAttribute('data-state', 'locked');

    const act = await consoleApi('GET', '/v2/console/base/activity?userId=' + P.id + '&kind=achievement');
    const actions = ((act.data && act.data.items) || []).map((x: any) => x.action);
    expect(actions).toEqual(expect.arrayContaining(['unlocked', 'revoked']));
    const audit = sql(`SELECT string_agg(DISTINCT action, ',' ORDER BY action) FROM console_audit_log WHERE message LIKE 'economy % users/${P.id}/achievements%'`);
    expect(audit.split(',')).toEqual(expect.arrayContaining(['CREATE', 'DELETE', 'INVOKE']));
    await expect(card.locator('[data-pac-log]')).toHaveAttribute('href', /kind=achievement/);
  });

  test('page search, English and 390px cards without sideways scroll', async ({ page }) => {
    await open(page, '#/economy/overview');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('دستاوردها');
    await expect(page.locator('.pl-pal-row').first()).toContainText('دستاوردها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/achievements/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/economy/achievements?q=' + key);
    const en = rows(page).filter({ hasText: 'Runner ' + k });
    await expect(en).toHaveCount(1);
    await expect(en).toContainText('Epic');
    await expect(en).toContainText('25 points');
    await expect(en.locator('[data-ach-sentence]')).toContainText('“Round finished” 2 times');
    await page.setViewportSize({ width: 390, height: 844 });
    await go(page, '#/economy/achievements');
    await noHorizontalScroll(page);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
  });
});
