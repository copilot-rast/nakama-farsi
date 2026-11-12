import { expect, test } from '../../lib/fixtures';
import { ADMIN, go, loginAs, open, root, settle, watchErrors } from '../../lib/console';

const PAGES: [string, string][] = [
  ['#/economy/overview', 'نمای کلی'],
  ['#/economy/currencies', 'ارزها'],
  ['#/economy/metas', 'متاها'],
  ['#/economy/prizes', 'جایزه‌ها'],
  ['#/economy/products', 'فروشگاه'],
  ['#/economy/gifts', 'هدیه‌ها'],
  ['#/economy/event-types', 'رویدادها'],
  ['#/economy/event-rules', 'جایزهٔ رویدادها'],
  ['#/economy/meta-rules', 'قانون‌های متا'],
  ['#/economy/level-tables', 'جدول سطح‌ها'],
  ['#/economy/streaks', 'روزهای پیاپی'],
  ['#/economy/tasks', 'تسک‌ها'],
  ['#/economy/task-groups', 'گروه‌های تسک'],
  ['#/economy/event-log', 'گزارش رویدادها'],
  ['#/economy/payments', 'پرداخت‌ها'],
  ['#/publication/faqs', 'سؤال‌های متداول'],
  ['#/publication/news', 'اخبار'],
  ['#/publication/comments', 'کامنت‌ها'],
  ['#/publication/notifications', 'اعلان‌ها'],
  ['#/moderation/players', 'مدیریت بازیکنان'],
  ['#/moderation/reports', 'گزارش‌ها'],
  ['#/moderation/restrictions', 'تعلیق و سکوت'],
  ['#/moderation/chat-words', 'فیلتر کلمات'],
];
const NATIVE_TABS = ['users', 'general', 'config', 'data-management', 'audit-log'];

test.describe('console shell', () => {
  test('sign in shows the Persian RTL console with our three sections and the project settings entry', async ({ spawn }) => {
    const page = await spawn({ signedIn: false });
    const errors = watchErrors(page);
    await page.goto('/#/login');
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
    await expect(page.locator('html')).toHaveAttribute('lang', 'fa-IR');
    await loginAs(page, ADMIN);
    const drawer = page.locator('.x-drawer');
    for (const [sec, title] of [['economy', 'اقتصاد'], ['publication', 'انتشار'], ['moderation', 'بازیکنان و گفتگو']]) {
      await expect(drawer.locator(`[data-economy-nav="${sec}"]`).first()).toContainText(title);
    }
    await expect(drawer.locator('[data-economy-nav="base"]')).toHaveCount(0);
    const order = await page.evaluate(() => [...document.querySelectorAll('.x-drawer [data-economy-nav]')].map((n) => n.getAttribute('data-economy-nav')));
    expect(order.slice(0, 3)).toEqual(['economy', 'publication', 'moderation']);
    await expect(drawer.locator('[data-ps-entry="sidebar"]').first()).toContainText('تنظیمات پروژه');
    await expect(page.locator('[data-ps-entry="topbar"]')).toHaveAttribute('aria-label', 'تنظیمات پروژه');
    await expect(page.locator('.sp-rail button[data-sp-group="economy"]')).toHaveAttribute('aria-label', 'اقتصاد');
    await expect(page.locator('.sp-rail button[data-sp-group="publication"]')).toHaveAttribute('aria-label', 'انتشار');
    await page.locator('.sp-rail button[data-sp-group="settings"]').click();
    await expect(page.locator('.sp-flyout.open [data-ps-entry="rail"]')).toContainText('تنظیمات پروژه');
    expect(errors).toEqual([]);
  });

  test('every section page opens from the rail with its heading and is the only active link', async ({ page }) => {
    await open(page, '#/economy/overview');
    for (const [hash, title] of PAGES) {
      const group = hash.split('/')[1];
      await page.locator(`.sp-rail button[data-sp-group="${group}"]`).click();
      await page.locator(`.sp-flyout.open a[href^="${hash}"]`).click();
      await expect(page).toHaveURL(new RegExp(hash.replace(/[/#?]/g, '\\$&')));
      await expect(root(page).locator('h1').first()).toHaveText(title);
      await settle(page);
      const active = page.locator('.sp-tabs a[aria-current="page"]');
      await expect(active).toHaveCount(1);
      await expect(active).toHaveAttribute('href', new RegExp('^' + hash.replace(/[/#?]/g, '\\$&')));
      await expect(page.locator('.sp-rail [aria-current]')).toHaveCount(1);
      await expect(page.locator(`.sp-rail button[data-sp-group="${group}"]`)).toHaveAttribute('aria-current', 'true');
      await expect(page.locator('.x-drawer [data-economy-nav] a[aria-current="page"]')).toHaveCount(1);
    }
    await page.locator('.sp-rail a.sp-ri[href="#/players"]').click();
    await expect(page).toHaveURL(/#\/players/);
    await expect(page.locator('#economy-root')).toHaveCount(0);
    await expect(page.locator('.sp-rail [aria-current]')).toHaveCount(1);
    await expect(page.locator('.sp-rail a.sp-ri[href="#/players"]')).toHaveAttribute('aria-current', 'page');
  });

  test('EN/FA switch re-renders our sections and pages in the other language', async ({ page }) => {
    await open(page, '#/economy/currencies');
    await expect(root(page).locator('h1')).toHaveText('ارزها');
    await page.locator('.sp-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(root(page).locator('h1')).toHaveText('Currencies');
    await expect(page.locator('.x-drawer [data-economy-nav="economy"]').first()).toContainText('Economy');
    await expect(page.locator('.x-drawer [data-economy-nav="publication"]').first()).toContainText('Publishing');
    await expect(page.locator('.x-drawer a[data-economy-page="news"]').first()).toContainText('News');
    await expect(page.locator('.sp-rail button[data-sp-group="economy"]')).toHaveAttribute('aria-label', 'Economy');
    await expect(page.locator('.sp-tabs a[aria-current="page"]')).toHaveText('Currencies');
    await go(page, '#/publication/news');
    await expect(root(page).locator('h1')).toHaveText('News');
    await page.locator('.sp-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('lang', 'fa-IR');
    await expect(root(page).locator('h1')).toHaveText('اخبار');
    await expect(page.locator('.x-drawer a[data-economy-page="news"]').first()).toContainText('اخبار');
    await expect(page.locator('.sp-tabs a[aria-current="page"]')).toHaveText('اخبار');
  });

  test('dark mode toggle persists over a reload and back', async ({ page }) => {
    await open(page, '#/economy/products');
    const html = page.locator('html');
    await expect(html).not.toHaveClass(/\bdark\b/);
    const toggle = page.locator('button.sp-theme');
    await expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await toggle.click();
    await expect(html).toHaveClass(/\bdark\b/);
    await expect(toggle).toHaveAttribute('aria-pressed', 'true');
    await page.reload();
    await page.locator('#economy-root').waitFor();
    await expect(html).toHaveClass(/\bdark\b/);
    await expect(page.locator('button.sp-theme')).toHaveAttribute('aria-pressed', 'true');
    const ink = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--ink-1').trim());
    expect(ink).toBe('#f1f2f6');
    await expect(page.locator('.sp-sky')).toHaveAttribute('data-theme', 'dark');
    await page.locator('button.sp-theme').click();
    await expect(html).not.toHaveClass(/\bdark\b/);
    await page.reload();
    await page.locator('#economy-root').waitFor();
    await expect(html).not.toHaveClass(/\bdark\b/);
  });

  test('reload and back keep our routes and their query', async ({ page }) => {
    await open(page, '#/economy/products');
    await go(page, '#/publication/news?isActive=true');
    await expect(root(page).locator('h1')).toHaveText('اخبار');
    await page.reload();
    await page.locator('#economy-root').waitFor();
    await settle(page);
    await expect(page).toHaveURL(/#\/publication\/news\?isActive=true/);
    await expect(root(page).locator('h1')).toHaveText('اخبار');
    await expect(root(page).locator('.eco-filters select').first()).toHaveValue('true');
    await page.goBack();
    await expect(page).toHaveURL(/#\/economy\/products/);
    await expect(root(page).locator('h1')).toHaveText('فروشگاه');
    await page.goForward();
    await expect(page).toHaveURL(/#\/publication\/news/);
    await expect(root(page).locator('h1')).toHaveText('اخبار');
  });

  test('legacy hashes redirect to the matching project settings section', async ({ page }) => {
    const cases: [string, string, string][] = [
      ['#/economy/categories', 'categories', ''],
      ['#/economy/tags', 'tags', ''],
      ['#/base/categories', 'categories', ''],
      ['#/base/tags?keyType=general', 'tags', '?keyType=general'],
      ['#/economy/payment-methods', 'payment-methods', ''],
      ['#/publication/providers', 'sms', ''],
      ['#/settings/payment-methods', 'payment-methods', ''],
      ['#/settings/sms', 'sms', ''],
      ['#/settings/email', 'email', ''],
      ['#/settings/languages', 'languages', ''],
      ['#/settings/brand', 'brand', ''],
      ['#/settings/setup', 'setup', ''],
    ];
    for (const [from, section, query] of cases) {
      await page.goto('/' + from);
      await expect(page).toHaveURL(new RegExp('#/project-settings/' + section + query.replace(/[?]/g, '\\$&') + '$'));
      await expect(page.locator(`.ps-window [role="tab"][data-ps-section="${section}"]`)).toHaveAttribute('aria-selected', 'true');
    }
    await open(page, '#/economy/overview');
    await go(page, '#/economy/categories?keyType=product');
    await expect(page).toHaveURL(/#\/project-settings\/categories\?keyType=product$/);
    await expect(root(page).locator('.eco-filters select').first()).toHaveValue('product');
  });

  test("Nakama's own settings keep their tabs and nothing of ours is injected", async ({ page }) => {
    await page.goto('/#/settings/users');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached' });
    for (const tab of NATIVE_TABS) await expect(page.locator(`a.x-tab[href="#/settings/${tab}"]`)).toHaveCount(1);
    await page.locator('a.x-tab[href="#/settings/general"]').click();
    await expect(page).toHaveURL(/#\/settings\/general$/);
    await expect(page.locator('a.x-tab[aria-current="page"]')).toHaveCount(1);
    await expect(page.locator('a.x-tab[data-settings-tab], .eco-tab')).toHaveCount(0);
    await expect(page.locator('#economy-root')).toHaveCount(0);
    await expect(page.locator('.ps-window')).toHaveCount(0);
  });
});
