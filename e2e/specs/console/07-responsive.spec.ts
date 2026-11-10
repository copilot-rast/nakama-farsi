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
  '#/project-settings/countries',
  '#/project-settings/accounts',
  '#/project-settings/otp',
  '#/project-settings/app-versions',
  '#/project-settings/app-versions/new',
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

const NATIVE_PHONE = ['#/', '#/players', '#/players/00000000-0000-0000-0000-000000000000', '#/players/00000000-0000-0000-0000-000000000000/wallet', '#/players/00000000-0000-0000-0000-000000000000/storage', '#/groups', '#/storage', '#/notifications', '#/runtime', '#/settings/users', '#/settings/config', '#/api-explorer'];

test.describe('native console at 390px with spatial', () => {
  for (const hash of NATIVE_PHONE) {
    test(`${hash} fits the phone, tables become labelled cards`, async ({ page }) => {
      await page.goto('/' + hash);
      await page.locator('.sp-menu').waitFor();
      await page.waitForFunction(() => !document.querySelector('#main .animate-pulse'));
      await page.waitForTimeout(300);
      await noHorizontalScroll(page);
      await expect(page.locator('.sp-rail')).toBeHidden();
      const table = page.locator('#main table[data-sp-cards]').first();
      if (await table.count()) {
        await expect(table.locator('thead')).toBeHidden();
        const cell = table.locator('tbody tr td[data-sp-label]').first();
        if (await cell.count()) {
          expect(await cell.evaluate((el) => getComputedStyle(el, '::before').content)).not.toBe('none');
          const box = await cell.boundingBox();
          expect(box!.width).toBeGreaterThan(200);
        }
      }
      for (const b of await page.locator('#main > div.sticky button:visible, #main > div.sticky a:visible').all()) {
        const r = await b.boundingBox();
        if (r) expect(Math.min(r.width, r.height)).toBeGreaterThanOrEqual(36);
      }
    });
  }

  test('the menu button opens the rail as a sheet from the start side, a group opens its pages, Escape closes', async ({ page }) => {
    await page.goto('/#/players');
    await page.locator('.sp-menu').click();
    const sheet = page.locator('.sp-rail-full');
    await expect(page.locator('html')).toHaveClass(/sp-sheet-open/);
    await expect(sheet).toBeVisible();
    const box = await sheet.boundingBox();
    expect(box!.x + box!.width).toBeGreaterThan(380);
    await sheet.locator('button[data-sp-group="economy"]').click();
    const fly = page.locator('.sp-flyout.open');
    await expect(fly).toBeVisible();
    const fb = await fly.boundingBox();
    expect(fb!.x).toBeGreaterThanOrEqual(0);
    expect(fb!.x + fb!.width).toBeLessThanOrEqual(390);
    await fly.locator('a[href^="#/economy/gifts"]').click();
    await expect(page).toHaveURL(/#\/economy\/gifts/);
    await expect(page.locator('html')).not.toHaveClass(/sp-sheet-open/);
    await page.locator('.sp-menu').click();
    await expect(page.locator('html')).toHaveClass(/sp-sheet-open/);
    await page.keyboard.press('Escape');
    await expect(page.locator('html')).not.toHaveClass(/sp-sheet-open/);
  });

  test('a native editor opens as a bottom sheet and closes', async ({ page }) => {
    await page.goto('/#/storage');
    await page.locator('#main tbody tr').first().waitFor();
    await page.locator('#main tbody tr').first().click();
    const form = page.locator('div.fixed.inset-0.flex.justify-end > .x-form');
    await expect(form).toBeVisible();
    await page.waitForFunction(() => document.getAnimations().every((a) => a.playState !== 'running' || a.effect?.getTiming().iterations === Infinity));
    const b = await form.boundingBox();
    expect(b!.width).toBeGreaterThanOrEqual(388);
    expect(Math.round(b!.y + b!.height)).toBeGreaterThanOrEqual(843);
    await noHorizontalScroll(page);
    await form.locator('button', { hasText: 'انصراف' }).click();
    await expect(form).toHaveCount(0);
  });
});
