import { expect, test } from '../../lib/fixtures';
import { go, noHorizontalScroll, open, panel, root, settle, settlePanel } from '../../lib/console';

const GROUPS = 'nakama-console:nav-groups';

const REDIRECTS: [string, string, string][] = [
  ['#/economy/meta-rules', '#/economy/metas/rules', 'meta-rules'],
  ['#/economy/level-tables', '#/economy/metas/levels', 'level-tables'],
  ['#/economy/task-groups', '#/economy/tasks/groups', 'task-groups'],
  ['#/economy/prize-groups', '#/economy/prizes/groups', 'prize-groups'],
  ['#/economy/event-types', '#/economy/events', 'event-types'],
  ['#/economy/event-rules', '#/economy/events/prizes', 'event-rules'],
  ['#/economy/event-log', '#/economy/events/report', 'event-log'],
  ['#/publication/banner-lists', '#/publication/banners/placements', 'banner-lists'],
  ['#/publication/notify-categories', '#/publication/notifications/categories', 'notify-categories'],
];

function exact(hash: string) {
  return new RegExp(hash.replace(/[/#?.]/g, '\\$&') + '$');
}

test.describe('navigation: sidebar groups, tabbed pages and old links', () => {
  test('economy groups fold, keep their state over a reload and open again for a page inside', async ({ page }) => {
    await open(page, '#/economy/overview');
    const heads = page.locator('.pl-side .pl-group-label');
    await expect(heads).toHaveText(['ارز و پیشرفت', 'فروش', 'جایزه و بازی', 'رویداد و رشد']);
    const sales = page.locator('.pl-side .pl-group-head[data-pl-group="sales"]');
    const shop = page.locator('.pl-side a.pl-nav[href^="#/economy/products"]');
    await expect(sales).toHaveAttribute('aria-expanded', 'true');
    await expect(shop).toBeVisible();
    await sales.click();
    await expect(sales).toHaveAttribute('aria-expanded', 'false');
    await expect(shop).toBeHidden();
    expect(await page.evaluate((k) => JSON.parse(localStorage.getItem(k) || '{}'), GROUPS)).toEqual({ 'economy/sales': false });
    await page.reload();
    await root(page).waitFor();
    await expect(sales).toHaveAttribute('aria-expanded', 'false');
    await expect(shop).toBeHidden();
    await expect(page.locator('.x-drawer [data-economy-group-toggle="sales"]').first()).toHaveAttribute('aria-expanded', 'false');
    await sales.press('Enter');
    await expect(shop).toBeVisible();
    await sales.click();
    await expect(shop).toBeHidden();
    await go(page, '#/economy/items');
    await expect(sales).toHaveAttribute('aria-expanded', 'true');
    await expect(page.locator('.pl-side a.pl-nav[aria-current="page"]')).toHaveText('آیتم‌ها');
    await expect(page.locator('.pl-tabs .pl-seg-item')).toHaveText(['فروشگاه', 'آیتم‌ها', 'هدیه‌ها', 'پرداخت‌ها']);
    await page.evaluate(() => { localStorage.setItem('nakama-console:nav-groups', '{oops'); });
    await page.reload();
    await root(page).waitFor();
    await expect(heads).toHaveCount(4);
    await expect(sales).toHaveAttribute('aria-expanded', 'true');
  });

  test('other sections are grouped the same way', async ({ page }) => {
    await open(page, '#/publication/faqs');
    await expect(page.locator('.pl-side .pl-group-label')).toHaveText(['محتوا', 'بنر و اسپانسر', 'پیام به بازیکن']);
    await expect(page.locator('.pl-side a.pl-nav')).toHaveText(['سؤال‌های متداول', 'اخبار', 'کامنت‌ها', 'بنرها', 'اسپانسرها', 'اعلان‌ها', 'اعلان‌های خودکار', 'پیکربندی بازی']);
    await go(page, '#/moderation/players');
    await expect(page.locator('.pl-side .pl-group-label')).toHaveText(['بازیکنان', 'گزارش و گفتگو']);
    await go(page, '#/support/tickets');
    await expect(page.locator('.pl-side .pl-group-head')).toHaveCount(0);
  });

  test('a tabbed page keeps its tab in the URL, over a reload and in the breadcrumbs', async ({ page }) => {
    await open(page, '#/economy/prizes');
    const tabs = root(page).locator('[data-economy-tabs="prizes"] a.eco-tab');
    await expect(tabs).toHaveText(['جایزه‌ها', 'گروه‌های جایزه']);
    await expect(tabs.first()).toHaveAttribute('aria-current', 'page');
    await tabs.nth(1).click();
    await expect(page).toHaveURL(exact('#/economy/prizes/groups'));
    await settle(page);
    await expect(root(page).locator('h1')).toHaveText('جایزه‌ها');
    await expect(tabs.nth(1)).toHaveAttribute('aria-current', 'page');
    await expect(page.locator('.pl-side a.pl-nav[aria-current="page"]')).toHaveText('جایزه‌ها');
    await expect(page.locator('.pl-crumbs .pl-crumb-link')).toHaveText(['اقتصاد', 'جایزه‌ها']);
    await expect(page.locator('.pl-crumbs .pl-crumb-current')).toHaveText('گروه‌های جایزه');
    await page.reload();
    await root(page).waitFor();
    await settle(page);
    await expect(tabs.nth(1)).toHaveAttribute('aria-current', 'page');
    await go(page, '#/economy/overview');
    await expect(page.locator('.pl-side a.pl-nav[href^="#/economy/prizes"]')).toHaveAttribute('href', '#/economy/prizes/groups');
    await go(page, '#/economy/prizes/groups/new');
    await settlePanel(page);
    await expect(tabs.nth(1)).toHaveAttribute('aria-current', 'page');
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);
    await expect(page).toHaveURL(exact('#/economy/prizes/groups'));
    await open(page, '#/economy/events/report?source=server');
    await expect(root(page).locator('[data-economy-tabs="events"] a.eco-tab')).toHaveText(['رویدادها', 'جایزهٔ رویدادها', 'گزارش']);
    await expect(root(page).locator('a[data-economy-tab="event-log"]')).toHaveAttribute('aria-current', 'page');
  });

  test('every old link opens its tab under the new address and keeps the rest of the link', async ({ page }) => {
    await open(page, '#/economy/overview');
    for (const [from, to, tab] of REDIRECTS) {
      await go(page, from);
      await expect(page).toHaveURL(exact(to));
      await expect(root(page).locator(`a[data-economy-tab="${tab}"]`)).toHaveAttribute('aria-current', 'page');
    }
    await open(page, '#/economy/event-types?q=login');
    await expect(page).toHaveURL(exact('#/economy/events?q=login'));
    await open(page, '#/economy/level-tables/new');
    await expect(page).toHaveURL(exact('#/economy/metas/levels/new'));
    await settlePanel(page);
    await expect(root(page).locator('a[data-economy-tab="level-tables"]')).toHaveAttribute('aria-current', 'page');
  });

  test('Ctrl K finds pages and the tabs inside them', async ({ page }) => {
    await open(page, '#/economy/overview');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('جایگاه');
    const row = page.locator('.pl-pal-row').first();
    await expect(row.locator('.pl-pal-label')).toHaveText('جایگاه‌های بنر');
    await expect(row.locator('.pl-pal-group')).toHaveText('انتشار · بنرها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(exact('#/publication/banners/placements'));
    await expect(page.locator('.pl-pal-input')).toHaveCount(0);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('قانون‌های متا');
    await expect(page.locator('.pl-pal-row').first().locator('.pl-pal-label')).toHaveText('قانون‌های متا');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(exact('#/economy/metas/rules'));
    await expect(page.locator('.pl-pal-input')).toHaveCount(0);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('آیتم');
    await expect(page.locator('.pl-pal-row').first().locator('.pl-pal-group')).toHaveText('اقتصاد · فروش');
  });

  test('English labels, and tabs and groups at 390px in dark mode without sideways scroll', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, locale: 'en', mode: 'dark', mobile: true });
    await open(page, '#/economy/events/prizes');
    await expect(root(page).locator('[data-economy-tabs="events"] a.eco-tab')).toHaveText(['Events', 'Event prizes', 'Report']);
    await noHorizontalScroll(page);
    await page.locator('.pl-menu').click();
    await expect(page.locator('.pl-side .pl-group-label').first()).toHaveText('Currency & progress');
    await expect(page.locator('.pl-side .pl-group-head[data-pl-group="growth"] .pl-group-label')).toHaveText('Events & growth');
    const box = await page.locator('.pl-side .pl-group-head').first().boundingBox();
    expect(box!.height).toBeGreaterThanOrEqual(36);
  });
});
