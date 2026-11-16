import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, root, toasts } from '../../lib/console';

test.describe('unity sdk guide', () => {
  test('header button opens the guide with the install line, copy and the steps', async ({ page }) => {
    await open(page, '#/economy/overview');
    const btn = page.locator('a[data-pl-sdk]');
    await expect(btn).toBeVisible();
    await expect(btn).toHaveText('پکیج یونیتی');
    await expect(btn).toHaveAttribute('title', 'راهنمای نصب پکیج یونیتی (SDK)');
    await btn.click();
    await expect(page).toHaveURL(/#\/system\/sdk$/);
    await expect(btn).toHaveAttribute('aria-current', 'page');
    const guide = root(page).locator('[data-sdk-page]');
    await expect(guide).toBeVisible();
    await expect(root(page).locator('h1')).toHaveText('پکیج یونیتی');
    await expect(guide.locator('[data-sdk-step]')).toHaveCount(8);
    await expect(guide.locator('[data-sdk-step="1"] .eco-sdk-n')).toHaveText('۱');
    await expect(guide.locator('[data-sdk-step="3"] pre')).toHaveText('"com.nakamafarsi.unity": "git@github.com:copilot-rast/nakama-farsi-unity.git?path=/Packages/Nakama#dev"');
    await expect(guide.locator('[data-sdk-step="1"] a')).toHaveAttribute('href', 'https://github.com/copilot-rast/nakama-farsi-unity');
    await expect(guide.locator('[data-sdk-step="6"]')).toContainText('Build for Bazaar');
    await expect(guide.locator('[data-sdk-step="5"] pre').first()).toContainText('NakamaFarsiRunner.Initialize');
    await guide.locator('[data-sdk-step="3"] [data-up-copy]').click();
    await expect(toasts(page).filter({ hasText: /کپی/ }).first()).toBeVisible();
  });

  test('the side menu and Ctrl K reach the guide', async ({ page }) => {
    await open(page, '#/system/updates');
    await expect(page.locator('.pl-side a.pl-nav[href="#/system/sdk"]')).toHaveText('پکیج یونیتی');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('یونیتی');
    await expect(page.locator('.pl-pal-row').first().locator('.pl-pal-label')).toHaveText('پکیج یونیتی');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/system\/sdk$/);
  });

  test('English, dark, 390px: icon-only header button and no sideways scroll', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/system/sdk');
    const btn = m.locator('a[data-pl-sdk]');
    await expect(btn).toBeVisible();
    await expect(btn.locator('.pl-sdk-label')).toBeHidden();
    await expect(btn).toHaveAttribute('aria-label', 'Unity SDK install guide');
    await expect(root(m).locator('h1')).toHaveText('Unity SDK');
    await expect(root(m).locator('[data-sdk-step="2"]')).toContainText('Unity 2021.3 or newer');
    await noHorizontalScroll(m);
  });
});
