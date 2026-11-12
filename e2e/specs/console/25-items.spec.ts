import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const key = 'e2e_' + k + '_skin';
const made = { users: [] as string[], items: [] as string[], prizes: [] as string[], products: [] as string[] };
let P: Player;

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

async function pickCoin(scope: Locator) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill('coin');
  const opt = pop.locator('[role="option"]').filter({ has: scope.page().locator('span', { hasText: /^coin$/ }) }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

async function confirm(page: Page, label: string) {
  await page.locator('[role="alertdialog"] button', { hasText: label }).click();
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-it-' + k, ('it' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  for (const id of made.products) await eco.del('/products/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  const all = await eco.get('/items?q=e2e_' + k + '&limit=100');
  for (const it of all.items || []) await eco.del('/items/' + it.id);
});

test.describe('items: console', () => {
  test('item editor: type, rarity, price, unlock level, upgrade rows, validation, list, search and locked key', async ({ page }) => {
    await open(page, '#/economy/items');
    await expect(root(page)).toContainText('چیزهایی که بازیکن در بازی دارد');
    await expect(rows(page).filter({ hasText: 'سبد یاقوتی' })).toHaveCount(1);
    await go(page, '#/economy/items/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(key);
    await fillTranslated(p, 'name', { fa: 'سبد آزمایشی ' + k, en: 'Test basket ' + k });
    await savePanel(page).catch(() => null);
    await expect(field(p, 'نوع آیتم')).toContainText('لازم');
    await pickIn(field(p, 'نوع آیتم'), 'پوسته');
    await field(p, 'کمیابی').locator('select').selectOption('epic');
    await pickCoin(field(p, 'ارز قیمت'));
    await savePanel(page).catch(() => null);
    await expect(field(p, /^قیمت$/)).toContainText('لازم');
    await field(p, /^قیمت$/).locator('input').fill('۲۵');
    await field(p, 'باز شدن از سطح').locator('input').fill('3');
    await p.locator('[data-rows-add="upgrades"]').click();
    await p.locator('[data-rows-add="upgrades"]').click();
    const up = p.locator('[data-rows="upgrades"] [data-row]');
    await expect(up).toHaveCount(2);
    await pickCoin(up.nth(0).locator('[data-col="currencyKey"]'));
    await pickCoin(up.nth(1).locator('[data-col="currencyKey"]'));
    await up.nth(1).locator('[data-col="amount"] input').fill('40');
    await savePanel(page);
    const item = await byKey('items', key);
    made.items.push(item.id);
    expect(item).toMatchObject({ typeKey: 'skin', rarity: 'epic', price: { currencyKey: 'coin', amount: 25 }, unlockLevel: 3, isStarter: false, isActive: true });
    expect(item.upgrades).toEqual([{ currencyKey: 'coin', amount: 10 }, { currencyKey: 'coin', amount: 40 }]);

    await search(page, key);
    await expect(rows(page)).toHaveCount(1);
    const row = rows(page).first();
    await expect(row).toContainText('سبد آزمایشی ' + k);
    await expect(row).toContainText('حماسی');
    await expect(row).toContainText('۲۵ سکه');
    await expect(row).toContainText('از سطح ۳');
    await expect(row).toContainText('۲ سطح ارتقا');
    await row.locator('td').first().click();
    await settlePanel(page);
    await expect(panel(page)).toContainText('ویرایش آیتم');
    await expect(field(panel(page), 'کلید').locator('input')).toHaveJSProperty('readOnly', true);
    await page.keyboard.press('Escape');

    await go(page, '#/economy/items/new');
    await settlePanel(page);
    const q = panel(page);
    await field(q, 'کلید').locator('input').fill(key + '_b');
    await fillTranslated(q, 'name', { fa: 'قاب آزمایشی ' + k });
    await pickIn(field(q, 'نوع آیتم'), 'قاب');
    await q.locator('.x-toggle', { hasText: 'همه از اول دارند' }).click();
    await expect(field(q, 'ارز قیمت')).toBeHidden();
    await pickIn(field(q, 'پیش‌نیاز'), 'سبد آزمایشی ' + k);
    await savePanel(page);
    const frame = await byKey('items', key + '_b');
    made.items.unshift(frame.id);
    expect(frame).toMatchObject({ typeKey: 'frame', isStarter: true, price: null, requiresKey: key });
    const used = await consoleApi('DELETE', '/v2/console/economy/items/' + item.id);
    expect(used.status).toBe(409);
    expect(used.data.error).toBe('IN_USE');
    await eco.put('/items/' + frame.id, Object.assign({}, frame, { isStarter: false, requiresKey: '', categoryId: '', levelTableId: '' }));
  });

  test('prize editor picks an item, product editor saves the purchase rules and the shop answers them', async ({ page }) => {
    await open(page, '#/economy/prizes/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_prize');
    await fillTranslated(p, 'name', { fa: 'جایزهٔ آیتم ' + k });
    const row = p.locator('.eco-items-row').first();
    await row.locator('select').first().selectOption('item');
    await pickIn(row.locator('[data-item-picker]'), 'سبد آزمایشی ' + k);
    await expect(row).toContainText('یک بار به بازیکن داده می‌شود');
    await savePanel(page);
    const prize = await byKey('prizes', 'e2e_' + k + '_prize');
    made.prizes.push(prize.id);
    expect(prize.items).toEqual([{ type: 'item', key, action: 'give', value: 1 }]);
    await search(page, 'e2e_' + k + '_prize');
    await expect(rows(page).first()).toContainText('آیتم: ' + key);

    await go(page, '#/economy/products/new');
    await settlePanel(page);
    const q = panel(page);
    await field(q, 'کلید').locator('input').fill('e2e_' + k + '_offer');
    await fillTranslated(q, 'name', { fa: 'پیشنهاد آیتم ' + k });
    await pickIn(field(q, 'جایزه'), 'جایزهٔ آیتم ' + k);
    await q.locator('.eco-choices label', { hasText: 'زرین‌پال' }).click();
    await q.locator('.eco-choices label', { hasText: 'سکهٔ بازی' }).click();
    await pickCoin(field(q, 'ارز'));
    await field(q, 'مقدار').locator('input').fill('5');
    await expect(q).toContainText('قانون‌های خرید');
    await field(q, 'حداقل سطح').locator('input').fill('2');
    await field(q, 'حداکثر سطح').locator('input').fill('1');
    await savePanel(page).catch(() => null);
    await expect(field(q, 'حداکثر سطح')).toContainText('کمتر');
    await field(q, 'حداکثر سطح').locator('input').fill('4');
    await q.locator('.x-toggle', { hasText: 'فقط یک بار خریدنی' }).click();
    await fillTranslated(q, 'purchaseMessage', { fa: 'نوش جانت ' + k, en: 'Enjoy ' + k });
    await savePanel(page);
    const product = await byKey('products', 'e2e_' + k + '_offer');
    made.products.push(product.id);
    expect(product).toMatchObject({ isOneTime: true, minLevel: 2, maxLevel: 4, checkContents: true, purchaseMessage: { fa: 'نوش جانت ' + k, en: 'Enjoy ' + k }, paymentMethods: ['in_app'] });
    await search(page, 'e2e_' + k + '_offer');
    await expect(rows(page).first().locator('[data-product-rules]')).toContainText('یک بار');
    await expect(rows(page).first().locator('[data-product-rules]')).toContainText('سطح ۲ تا ۴');

    const cat = await rpc(P, 'economy_catalog', { lang: 'fa' });
    const offer = cat.data.products.find((x: any) => x.key === 'e2e_' + k + '_offer');
    expect(offer).toMatchObject({ isOneTime: true, levelLocked: true, owned: false, bought: false, purchaseCount: 0, purchaseMessage: 'نوش جانت ' + k, items: [key] });
    const low = await rpc(P, 'economy_purchase_start', { productId: product.id, method: 'in_app' });
    expect(low.data.message || low.data).toBeTruthy();
    expect(JSON.stringify(low.data)).toContain('LEVEL_TOO_LOW');
  });

  test('player inventory card: starter items, give an item, take it back, and the activity log', async ({ page }) => {
    const before = await consoleApi('GET', '/v2/console/economy/users/' + P.id + '/items?lang=fa');
    expect(before.data.items.find((x: any) => x.key === 'star_basket_classic')).toMatchObject({ owned: true, equipped: true, source: 'starter' });
    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-items]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card).toContainText('انبار بازیکن');
    await expect(card.locator('[data-inv-item="star_basket_classic"]')).toContainText('از اول');
    await expect(card.locator('[data-inv-revoke="star_basket_classic"]')).toHaveCount(0);
    await pickIn(card.locator('[data-inv-give]'), 'سبد آزمایشی ' + k);
    await card.locator('[data-inv-grant]').click();
    await expect(toasts(page).last()).toContainText('آیتم به بازیکن داده شد');
    const line = card.locator(`[data-inv-item="${key}"]`);
    await expect(line).toContainText('از کنسول');
    await expect(line).toContainText('ارتقای ۰ از ۲');
    const inv = await rpc(P, 'economy_inventory', { lang: 'fa' });
    expect(inv.data.items.find((x: any) => x.key === key)).toMatchObject({ owned: true, source: 'admin' });
    await line.locator(`[data-inv-revoke="${key}"]`).click();
    await confirm(page, 'پس گرفتن');
    await expect(toasts(page).last()).toContainText('آیتم پس گرفته شد');
    await expect(card.locator(`[data-inv-item="${key}"]`)).toHaveCount(0);
    const act = await consoleApi('GET', '/v2/console/base/activity?userId=' + P.id + '&kind=item');
    const actions = ((act.data && act.data.items) || []).map((x: any) => x.action);
    expect(actions).toEqual(expect.arrayContaining(['granted', 'revoked', 'starter']));
  });

  test('page search, English and 390px cards without sideways scroll', async ({ page }) => {
    await open(page, '#/economy/overview');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('آیتم‌ها');
    await expect(page.locator('.pl-pal-row').first()).toContainText('آیتم‌ها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/items/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/economy/items?q=' + key);
    const en = rows(page).filter({ hasText: 'Test basket ' + k });
    await expect(en).toHaveCount(1);
    await expect(en).toContainText('Epic');
    await expect(en).toContainText('From level 3');
    await page.setViewportSize({ width: 390, height: 844 });
    await go(page, '#/economy/items');
    await noHorizontalScroll(page);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
  });
});
