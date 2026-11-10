import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, root } from '../../lib/console';

test.use({ pageOpts: { signedIn: true, mobile: true } } as any);

const PAGES = [
  '#/project-settings/categories',
  '#/project-settings/tags',
  '#/economy/overview',
  '#/economy/currencies',
  '#/economy/prizes',
  '#/economy/products',
  '#/economy/gifts',
  '#/economy/payments',
  '#/publication/faqs',
  '#/publication/news',
  '#/publication/comments',
  '#/publication/notifications',
  '#/publication/notifications/new',
  '#/project-settings/payment-methods',
  '#/project-settings/sms',
  '#/project-settings/email',
  '#/project-settings/languages',
  '#/project-settings/brand',
  '#/project-settings/setup',
];

test.describe('console at 390px', () => {
  for (const hash of PAGES) {
    test(`${hash} fits the phone and lists become cards`, async ({ page }) => {
      await open(page, hash);
      await expect(root(page)).toBeVisible();
      await noHorizontalScroll(page);
      const tables = root(page).locator('table.eco-table');
      for (let i = 0; i < (await tables.count()); i++) {
        const t = tables.nth(i);
        await expect(t.locator('thead')).toBeHidden();
        const row = t.locator('tbody tr').first();
        if (await row.count()) {
          expect(await row.evaluate((el) => getComputedStyle(el).display)).toBe('block');
          const label = await row.locator('td').nth(1).evaluate((el) => getComputedStyle(el, '::before').content).catch(() => 'none');
          expect(label).not.toBe('none');
        }
      }
    });
  }

  test('editor drawer and setup wizard fit the phone', async ({ page }) => {
    await open(page, '#/economy/products/new');
    const panel = page.locator('[data-economy="panel"] .x-form');
    await expect(panel).toBeVisible();
    const box = await panel.boundingBox();
    expect(box!.width).toBeLessThanOrEqual(390);
    await noHorizontalScroll(page);
    await page.keyboard.press('Escape');
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    const wiz = page.locator('[data-economy="setup"]');
    await expect(wiz).toBeVisible();
    const steps = wiz.locator('.setup-step');
    await expect(steps).toHaveCount(7);
    const wb = await wiz.boundingBox();
    expect(wb!.width).toBeLessThanOrEqual(390);
    const overflow = await wiz.evaluate((el) => {
      const body = el.querySelector('.setup-body') as HTMLElement;
      return body ? body.scrollWidth - body.clientWidth : 0;
    });
    expect(overflow).toBeLessThanOrEqual(1);
    await wiz.locator('[data-setup-later]').click();
    await expect(wiz).toHaveCount(0);
  });
});
