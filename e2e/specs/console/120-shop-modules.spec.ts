import { expect, test } from '../../lib/fixtures';
import { base, eco, uid } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { field, go, noHorizontalScroll, open, openSettings, panel, psWindow, root, rows, settle, settlePanel, toasts } from '../../lib/console';

const REDIRECTS: [string, string][] = [
  ['#/economy/products', '#/economy/shop/products'],
  ['#/economy/offers', '#/economy/shop/offers'],
  ['#/economy/offers/entry', '#/economy/shop/offers/entry'],
  ['#/economy/entry-offers', '#/economy/shop/offers/entry'],
  ['#/economy/offers/report', '#/economy/shop/offers/report'],
  ['#/economy/offer-preview', '#/economy/shop/offers/preview'],
  ['#/economy/tag-discounts', '#/economy/shop/discounts'],
  ['#/economy/tag-discounts/preview', '#/economy/shop/discounts/preview'],
  ['#/economy/items', '#/economy/shop/items'],
  ['#/economy/avatar', '#/economy/shop/avatar'],
  ['#/economy/avatar/types', '#/economy/shop/avatar/types'],
  ['#/economy/avatar-items', '#/economy/shop/avatar/items'],
  ['#/economy/gifts', '#/economy/shop/gifts'],
  ['#/economy/friend-gifts', '#/economy/shop/gifts/friends'],
  ['#/economy/friend-gifts/rules', '#/economy/shop/gifts/friends/rules'],
  ['#/economy/friend-gift-report', '#/economy/shop/gifts/friends/report'],
  ['#/economy/payments', '#/economy/shop/payments'],
];

function exact(hash: string) {
  return new RegExp(hash.replace(/[/#?.]/g, '\\$&') + '$');
}

async function setModules(disabled: string[]) {
  await base.put('/console-modules', { disabled });
}

async function palette(page: import('@playwright/test').Page, text: string) {
  await page.locator('.pl-side-search').click();
  await page.locator('.pl-pal-input:visible').fill(text);
}

test.describe('shop hub: one menu link, tabs, second level and old links', () => {
  test('the sales group is one «فروشگاه» link and the shop has seven tabs', async ({ page }) => {
    await open(page, '#/economy/overview');
    await expect(page.locator('.pl-side .pl-group-label')).toHaveText(['ارز و پیشرفت', 'جایزه و بازی', 'رویداد و رشد']);
    const shop = page.locator('.pl-side a.pl-nav[href^="#/economy/shop"]');
    await expect(shop).toHaveText('فروشگاه');
    await shop.click();
    await expect(page).toHaveURL(exact('#/economy/shop/products'));
    await settle(page);
    await expect(root(page).locator('h1')).toHaveText('فروشگاه');
    await expect(root(page).locator('[data-economy-tabs="shop"] a.eco-tab')).toHaveText(['محصولات', 'پیشنهادها', 'تخفیف‌ها', 'آیتم‌ها', 'آواتار', 'هدیه‌ها', 'پرداخت']);
    await expect(root(page).locator('[data-shop-tab="products"]')).toHaveAttribute('aria-current', 'page');
    await expect(root(page).locator('[data-shop-sub]')).toHaveCount(0);
    await expect(page.locator('.pl-side a.pl-nav[aria-current="page"]')).toHaveText('فروشگاه');
    await expect(page.locator('.pl-crumbs .pl-crumb-current')).toHaveText('محصولات');
  });

  test('every old link lands on its shop tab', async ({ page }) => {
    await open(page, '#/economy/overview');
    for (const [from, to] of REDIRECTS) {
      await go(page, from);
      await expect(page, from).toHaveURL(exact(to));
      await expect(root(page).locator('[data-economy-tabs="shop"]')).toBeVisible();
    }
  });

  test('sibling lists sit in a segmented control and tools at the end of the same row', async ({ page }) => {
    await open(page, '#/economy/shop/offers');
    const sub = root(page).locator('[data-shop-sub="offers"]');
    await expect(sub.locator('[data-shop-list]')).toHaveText(['روزانه', 'پیشنهاد ورود']);
    await expect(sub.locator('[data-shop-tool]')).toHaveText(['پیش‌نمایش بازیکن', 'گزارش', 'تنظیمات']);
    await sub.locator('[data-shop-list="entry-offers"]').click();
    await expect(page).toHaveURL(exact('#/economy/shop/offers/entry'));
    await settle(page);
    await expect(sub.locator('[data-shop-list="entry-offers"]')).toHaveAttribute('aria-selected', 'true');
    await sub.locator('[data-shop-tool="offer-report"]').click();
    await expect(page).toHaveURL(exact('#/economy/shop/offers/report'));
    await settle(page);
    await expect(sub.locator('[data-shop-back]')).toHaveText('بازگشت به پیشنهاد ورود');
    await expect(sub.locator('[data-shop-tool="offer-report"]')).toHaveAttribute('aria-current', 'page');
    await sub.locator('[data-shop-back]').click();
    await expect(page).toHaveURL(exact('#/economy/shop/offers/entry'));
    await settle(page);
    await go(page, '#/economy/shop/gifts');
    const gifts = root(page).locator('[data-shop-sub="gifts"]');
    await expect(gifts.locator('[data-shop-list]')).toHaveText(['کدهای هدیه', 'هدیه به دوست']);
    await expect(gifts.locator('[data-shop-tool]')).toHaveCount(0);
    await gifts.locator('[data-shop-list="friend-gifts"]').click();
    await settle(page);
    await expect(gifts.locator('[data-shop-tool]')).toHaveText(['قانون‌ها', 'گزارش', 'تنظیمات']);
    await go(page, '#/economy/shop/avatar');
    await expect(root(page).locator('[data-shop-sub="avatar"] [data-shop-list]')).toHaveText(['دسته‌ها', 'آیتم‌ها', 'نوع‌ها']);
    await go(page, '#/economy/shop/payments');
    await expect(root(page).locator('[data-shop-sub="payments"] [data-shop-tool]')).toHaveText(['درگاه‌ها', 'تنظیمات']);
  });

  test('settings open in the side panel over the list and Esc goes back', async ({ page }) => {
    await open(page, '#/economy/shop/offers');
    await root(page).locator('[data-shop-tool="offer-settings"]').click();
    await expect(page).toHaveURL(exact('#/economy/shop/offers/settings'));
    const sheet = page.locator('[data-shop-settings="offer-settings"]');
    await expect(sheet.getByRole('dialog')).toBeVisible();
    await expect(sheet.getByRole('dialog')).toHaveAccessibleName('تنظیمات پیشنهادها');
    await expect(root(page).locator('[data-shop-list="offers"]')).toHaveAttribute('aria-selected', 'true');
    await expect(sheet.locator('.animate-pulse')).toHaveCount(0);
    await page.reload();
    await expect(page.locator('[data-shop-settings="offer-settings"]').getByRole('dialog')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page).toHaveURL(exact('#/economy/shop/offers'));
    await expect(page.locator('[data-shop-settings]')).toHaveCount(0);
    await go(page, '#/economy/payments/settings');
    await expect(page).toHaveURL(exact('#/economy/shop/payments/settings'));
    await expect(page.locator('[data-shop-settings="payment-settings"]').getByRole('dialog')).toBeVisible();
  });

  test('Ctrl K still finds every page inside the shop', async ({ page }) => {
    await open(page, '#/economy/overview');
    await palette(page, 'قانون‌های هدیه به دوست');
    await expect(page.locator('.pl-pal-row').first()).toContainText('قانون‌های هدیه به دوست');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(exact('#/economy/shop/gifts/friends/rules'));
    await palette(page, 'گزارش پیشنهادها');
    await expect(page.locator('.pl-pal-row').first()).toContainText('گزارش پیشنهادها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(exact('#/economy/shop/offers/report'));
  });

  test('products: category filter, row menu, copy and mobile cards', async ({ page, spawn }) => {
    const cur = await eco.post('/currencies', { key: uid('cur'), name: { fa: 'سکهٔ آزمون' }, type: 'soft', defaultValue: 0 });
    const prize = await eco.post('/prizes', { key: uid('pz'), name: { fa: 'جایزهٔ آزمون' }, message: {}, tagIds: [], isRepeatable: true, items: [{ type: 'currency', key: cur.key, action: 'add', value: 10 }], metadata: {} });
    const key = uid('pr');
    const product = await eco.post('/products', { key, isActive: true, name: { fa: 'محصول آزمون ' + key }, description: {}, tagIds: [], audienceTagIds: [], prizeId: prize.id, paymentMethods: ['zarinpal'], priceToman: 12000, discountPercent: 0, sku: {}, coinPrice: null, isSubscription: false, minLevel: 0, sort: 0, metadata: {} });
    try {
      await open(page, '#/economy/shop/products?q=' + key);
      await expect(root(page).locator('[data-filter="categoryId"]')).toBeVisible();
      await expect(rows(page)).toHaveCount(1);
      const btn = rows(page).first().locator('.eco-rowmenu-btn');
      await rows(page).first().hover();
      await btn.click();
      const menu = rows(page).first().getByRole('menu');
      await expect(menu.getByRole('menuitem')).toHaveText(['ویرایش', 'ساخت کپی', 'دیدن خریدها', 'غیرفعال کردن', 'حذف']);
      await expect(menu.getByRole('menuitem').first()).toBeFocused();
      await page.keyboard.press('Escape');
      await expect(menu).toBeHidden();
      await btn.click();
      await menu.locator('[data-row-action="active"]').click();
      await expect(toasts(page)).toContainText('محصول غیرفعال شد.');
      await expect.poll(async () => (await eco.get('/products/' + product.id)).isActive).toBe(false);
      await btn.click();
      await menu.locator('[data-row-action="copy"]').click();
      await expect(page).toHaveURL(new RegExp('#/economy/shop/products/new\\?.*copy=' + product.id));
      await settlePanel(page);
      await expect(field(panel(page), 'کلید').locator('input')).toHaveValue(key + '_copy');
      await page.keyboard.press('Escape');
      const phone = await spawn({ mobile: true, signedIn: true });
      await open(phone, '#/economy/shop/products?q=' + key);
      await expect(rows(phone)).toHaveCount(1);
      await expect(rows(phone).first().locator('.eco-rowmenu-btn')).toBeVisible();
      await noHorizontalScroll(phone);
    } finally {
      await eco.del('/products/' + product.id);
      await eco.del('/prizes/' + prize.id);
      await eco.del('/currencies/' + cur.id);
    }
  });
});

test.describe('modules: switch off what the game does not use', () => {
  test.afterAll(() => { restoreConfig(); });

  test('turning tournaments off hides the rail mode, Ctrl K and shows the off page on a direct link', async ({ page }) => {
    await setModules([]);
    await openSettings(page, 'modules');
    const win = psWindow(page);
    await expect(win.locator('[data-mod-group]')).toHaveCount(7);
    const sw = win.locator('[data-mod-switch="cups"]');
    await expect(sw).toHaveAttribute('aria-checked', 'true');
    await sw.click();
    await expect(sw).toHaveAttribute('aria-checked', 'false');
    await expect(win.locator('[data-mod-note="cups"]')).toBeVisible();
    await expect(toasts(page)).toContainText('تورنمنت‌ها خاموش شد.');
    expect((await base.get('/console-modules')).disabled).toEqual(['cups']);
    await page.locator('[data-toast-undo]').click();
    await expect(sw).toHaveAttribute('aria-checked', 'true');
    expect((await base.get('/console-modules')).disabled).toEqual([]);
    await sw.click();
    await expect(sw).toHaveAttribute('aria-checked', 'false');
    await win.locator('[data-mod-filter="off"]').click();
    await expect(win.locator('[data-mod-row]')).toHaveCount(1);
    await page.keyboard.press('Escape');
    await expect(page.locator('.pl-rail [data-pl-mode="cups"]')).toHaveCount(0);
    await expect(page.locator('.x-drawer [data-cup-group]')).toHaveCount(0);
    await palette(page, 'فهرست تورنمنت‌ها');
    await expect(page.locator('.pl-pal-row').filter({ hasText: 'فهرست تورنمنت‌ها' })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await page.goto('/#/cups/list');
    const off = page.locator('[data-mod-off="cups"]');
    await expect(off).toBeVisible();
    await expect(off.locator('h2')).toHaveText('ماژول تورنمنت‌ها خاموش است');
    await expect(page.locator('.pl-crumbs .pl-crumb-current')).toHaveText('تورنمنت‌ها');
    await off.locator('[data-mod-turn-on="cups"]').click();
    await expect(off).toHaveCount(0);
    await expect(page.locator('.pl-rail [data-pl-mode="cups"]')).toHaveCount(1);
    expect((await base.get('/console-modules')).disabled).toEqual([]);
  });

  test('an economy module off drops its shop tab, its menu link, its project settings section and its overview card', async ({ page }) => {
    await setModules(['offers', 'events', 'support']);
    await open(page, '#/economy/shop/products');
    await expect(root(page).locator('[data-economy-tabs="shop"] a.eco-tab')).toHaveText(['محصولات', 'تخفیف‌ها', 'آیتم‌ها', 'آواتار', 'هدیه‌ها', 'پرداخت']);
    await expect(page.locator('.pl-side a.pl-nav').filter({ hasText: 'رویدادها' })).toHaveCount(0);
    await expect(page.locator('.pl-rail [data-pl-mode="support"]')).toHaveCount(0);
    await go(page, '#/economy/shop/offers');
    await expect(root(page).locator('[data-mod-off="offers"]')).toBeVisible();
    await go(page, '#/economy/overview');
    await expect(root(page).locator('[data-ov-tile="offers"], [data-ov-tile="event-types"]')).toHaveCount(0);
    await openSettings(page, 'brand');
    await expect(psWindow(page).locator('[role="tab"][data-ps-section="events"]')).toHaveCount(0);
    await expect(psWindow(page).locator('[role="tab"][data-ps-section="support-settings"]')).toHaveCount(0);
    await setModules([]);
  });

  test('the modules page at 390px fits the screen', async ({ spawn }) => {
    await setModules(['cups']);
    const phone = await spawn({ mobile: true, signedIn: true });
    await openSettings(phone, 'modules');
    await expect(psWindow(phone).locator('[data-mod-row="cups"]')).toHaveClass(/eco-mod-row--off/);
    await noHorizontalScroll(phone);
    await setModules([]);
  });
});
