import { expect, test } from '../../lib/fixtures';
import { ADMIN, go, loginAs, open, root, settle, watchErrors } from '../../lib/console';

const PAGES: [string, string][] = [
  ['#/base/categories', 'دسته‌بندی‌ها'],
  ['#/base/tags', 'تگ‌ها'],
  ['#/economy/overview', 'نمای کلی'],
  ['#/economy/currencies', 'ارزها'],
  ['#/economy/metas', 'متاها'],
  ['#/economy/prizes', 'جایزه‌ها'],
  ['#/economy/products', 'فروشگاه'],
  ['#/economy/gifts', 'هدیه‌ها'],
  ['#/economy/payments', 'پرداخت‌ها'],
  ['#/publication/faqs', 'سؤال‌های متداول'],
  ['#/publication/news', 'اخبار'],
  ['#/publication/comments', 'کامنت‌ها'],
  ['#/publication/notifications', 'اعلان‌ها'],
];
const TABS = ['payment-methods', 'sms', 'email', 'languages', 'brand', 'setup'];

test.describe('console shell', () => {
  test('sign in shows the Persian RTL console with our three sections', async ({ spawn }) => {
    const page = await spawn({ signedIn: false });
    const errors = watchErrors(page);
    await page.goto('/#/login');
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
    await expect(page.locator('html')).toHaveAttribute('lang', 'fa-IR');
    await loginAs(page, ADMIN);
    const drawer = page.locator('.x-drawer');
    for (const [sec, title] of [['base', 'عمومی'], ['economy', 'اقتصاد'], ['publication', 'انتشار']]) {
      await expect(drawer.locator(`[data-economy-nav="${sec}"]`).first()).toContainText(title);
    }
    const order = await page.evaluate(() => [...document.querySelectorAll('.x-drawer [data-economy-nav]')].map((n) => n.getAttribute('data-economy-nav')));
    expect(order.slice(0, 3)).toEqual(['base', 'economy', 'publication']);
    expect(errors).toEqual([]);
  });

  test('every sidebar page opens with its heading and is the only active link', async ({ page }) => {
    await open(page, '#/economy/overview');
    for (const [hash, title] of PAGES) {
      await page.locator(`.x-drawer a[data-economy-page][href^="${hash}"]`).first().click();
      await expect(page).toHaveURL(new RegExp(hash.replace(/[/#?]/g, '\\$&')));
      await expect(root(page).locator('h1').first()).toHaveText(title);
      await settle(page);
      const active = page.locator('.x-drawer [data-economy-nav] a[aria-current="page"]');
      await expect(active).toHaveCount(1);
      await expect(active).toHaveAttribute('href', new RegExp('^' + hash.replace(/[/#?]/g, '\\$&')));
      await expect(page.locator('.x-drawer a[href="#/"][aria-current="page"]:not([data-economy-dimmed])')).toHaveCount(0);
    }
    await page.locator('.x-drawer a[href="#/players"]').first().click();
    await expect(page).toHaveURL(/#\/players/);
    await expect(page.locator('#economy-root')).toHaveCount(0);
    await expect(page.locator('.x-drawer [data-economy-nav] a[aria-current="page"]')).toHaveCount(0);
  });

  test('EN/FA switch re-renders our sections and pages in the other language', async ({ page }) => {
    await open(page, '#/economy/currencies');
    await expect(root(page).locator('h1')).toHaveText('ارزها');
    await page.locator('.l10n-switcher:not(.economy-theme)').click();
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(root(page).locator('h1')).toHaveText('Currencies');
    await expect(page.locator('.x-drawer [data-economy-nav="economy"]').first()).toContainText('Economy');
    await expect(page.locator('.x-drawer [data-economy-nav="publication"]').first()).toContainText('Publishing');
    await expect(page.locator('.x-drawer a[data-economy-page="news"]').first()).toContainText('News');
    await go(page, '#/publication/news');
    await expect(root(page).locator('h1')).toHaveText('News');
    await page.locator('.l10n-switcher:not(.economy-theme)').click();
    await expect(page.locator('html')).toHaveAttribute('lang', 'fa-IR');
    await expect(root(page).locator('h1')).toHaveText('اخبار');
    await expect(page.locator('.x-drawer a[data-economy-page="news"]').first()).toContainText('اخبار');
  });

  test('dark mode toggle persists over a reload and back', async ({ page }) => {
    await open(page, '#/economy/products');
    const html = page.locator('html');
    await expect(html).not.toHaveClass(/\bdark\b/);
    const toggle = page.locator('button.economy-theme');
    await expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await toggle.click();
    await expect(html).toHaveClass(/\bdark\b/);
    await expect(toggle).toHaveAttribute('aria-pressed', 'true');
    await page.reload();
    await page.locator('#economy-root').waitFor();
    await expect(html).toHaveClass(/\bdark\b/);
    await expect(page.locator('button.economy-theme')).toHaveAttribute('aria-pressed', 'true');
    const bg = await page.locator('#economy-root table').first().evaluate((el) => getComputedStyle(el.closest('.bg-white') as Element).backgroundColor);
    expect(bg).not.toBe('rgb(255, 255, 255)');
    await page.locator('button.economy-theme').click();
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

  test('legacy hashes redirect to where the pages live now', async ({ page }) => {
    const cases: [string, RegExp, string][] = [
      ['#/economy/categories', /#\/base\/categories$/, 'h1'],
      ['#/economy/tags', /#\/base\/tags$/, 'h1'],
      ['#/economy/payment-methods', /#\/settings\/payment-methods$/, 'payment-methods'],
      ['#/publication/providers', /#\/settings\/sms$/, 'sms'],
    ];
    for (const [from, to, what] of cases) {
      await page.goto('/' + from);
      await expect(page).toHaveURL(to);
      await page.locator('#economy-root').waitFor();
      if (what === 'h1') await expect(root(page).locator('h1')).toHaveText(from.includes('tags') ? 'تگ‌ها' : 'دسته‌بندی‌ها');
      else await expect(page.locator(`a.x-tab[data-settings-tab="${what}"]`)).toHaveAttribute('aria-current', 'page');
    }
    await open(page, '#/economy/overview');
    await go(page, '#/economy/categories?keyType=product');
    await expect(page).toHaveURL(/#\/base\/categories\?keyType=product$/);
  });

  test('settings tabs show exactly one active tab at a time', async ({ page }) => {
    await open(page, '#/settings/payment-methods');
    for (const tab of TABS) {
      await page.locator(`a.x-tab[data-settings-tab="${tab}"]`).click();
      await expect(page).toHaveURL(new RegExp('#/settings/' + tab + '$'));
      await page.locator('#economy-root[data-economy="settings"]').waitFor();
      await expect(page.locator('a.x-tab[aria-current="page"]')).toHaveCount(1);
      await expect(page.locator('a.x-tab.eco-tab-on')).toHaveCount(1);
      await expect(page.locator('a.x-tab[aria-current="page"]')).toHaveAttribute('data-settings-tab', tab);
      await expect(page.locator('#economy-root')).toHaveCount(1);
    }
    await page.locator('a.x-tab[href="#/settings/general"]').click();
    await expect(page).toHaveURL(/#\/settings\/general$/);
    await expect(page.locator('#economy-root')).toHaveCount(0);
    await expect(page.locator('a.x-tab.eco-tab-on')).toHaveCount(0);
    await expect(page.locator('a.x-tab[data-settings-tab][aria-current="page"]')).toHaveCount(0);
    await page.locator('a.x-tab[data-settings-tab="languages"]').click();
    await expect(page.locator('a.x-tab[aria-current="page"]')).toHaveCount(1);
    await expect(page.locator('a.x-tab[href="#/settings/general"]')).not.toHaveAttribute('aria-current', 'page');
  });
});
