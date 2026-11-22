import fs from 'node:fs';
import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { base } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { ADMIN } from '../../lib/console';
import { CONSOLE_URL, FIXTURES } from '../../lib/env';
import type { Page } from '@playwright/test';

const MARK = 'https://e2e-login.invalid/mark.svg';
const CARD = '#app .sm\\:w-\\[512px\\]';

test.describe.configure({ mode: 'serial' });
test.afterEach(() => restoreConfig());

async function brandPage(page: Page) {
  await page.route(MARK, (r) => r.fulfill({ status: 200, contentType: 'image/svg+xml', body: fs.readFileSync(path.join(FIXTURES, 'mark.svg')) }));
  await page.goto(CONSOLE_URL + '/#/login');
  await page.locator(`${CARD} [data-brand-signin]`).waitFor();
}

test.describe('sign-in card', () => {
  test('logo, title and toggles sit inside the card; Persian digits; theme toggle; sign-in still stores the session', async ({ spawn }) => {
    await base.put('/brand', { name: { fa: 'گشت پلیس 3', en: 'Police Patrol 3' }, shortName: 'PP3', logoUrl: '', logoDarkUrl: '', markUrl: MARK, accent: '#6b59ee', welcome: {} });
    const page = await spawn({ signedIn: false, mode: 'dark' });
    await brandPage(page);
    const card = page.locator(CARD);
    const mark = card.locator('[data-brand-signin] [data-brand-mark]');
    await expect(mark.locator('img')).toBeVisible();
    await expect(mark.locator('img')).toHaveAttribute('data-fit', 'cover');
    expect(parseFloat(await mark.evaluate((e) => getComputedStyle(e).borderTopLeftRadius))).toBeGreaterThanOrEqual(12);
    const title = card.locator('.brand-signin-title');
    await expect(title).toHaveText('گشت پلیس ۳');
    await expect(card.locator('.brand-signin-sub')).toHaveText('برای ادامه وارد شوید');
    const lines = await title.evaluate((e) => Math.round(e.getBoundingClientRect().height / parseFloat(getComputedStyle(e).lineHeight)));
    expect(lines).toBe(1);
    const [m, t] = [await mark.boundingBox(), await title.boundingBox()];
    expect(m!.y + m!.height).toBeLessThanOrEqual(t!.y);
    await expect(page.locator('#app a[href="/"]:visible')).toHaveCount(0);
    await expect(page.locator('body > .l10n-switcher:visible')).toHaveCount(0);

    const tools = card.locator('[data-signin-tools]');
    await expect(tools.locator('[data-locale="fa-IR"]')).toHaveAttribute('aria-checked', 'true');
    await expect(tools.locator('[data-locale="en"]')).toHaveAttribute('aria-checked', 'false');
    const theme = tools.locator('[data-signin-theme]');
    await expect(theme).toHaveAttribute('aria-pressed', 'true');
    await theme.click();
    await expect(page.locator('html')).not.toHaveClass(/\bdark\b/);
    await expect(theme).toHaveAttribute('aria-pressed', 'false');
    expect(await page.evaluate(() => localStorage.getItem('color-mode'))).toBe('light');

    const user = page.locator('input[name="username"]');
    await expect(async () => {
      await user.fill(ADMIN.user);
      await page.locator('input[name="password"]').fill(ADMIN.pass);
      await expect(user).toHaveValue(ADMIN.user);
    }).toPass();
    await page.keyboard.press('Enter');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
    await expect(page.locator('[data-brand-signin], [data-signin-tools]')).toHaveCount(0);
    await expect(page.locator('html')).not.toHaveClass(/brand-signin-on/);
    expect(await page.evaluate(() => !!(sessionStorage.getItem('SESSION_TOKENS') || localStorage.getItem('SESSION_TOKENS')))).toBe(true);
  });

  test('language switch in the card turns the page English and LTR at 390px; no brand shows the neutral mark', async ({ spawn }) => {
    await base.put('/brand', {});
    const page = await spawn({ signedIn: false, mobile: true });
    await page.goto(CONSOLE_URL + '/#/login');
    const card = page.locator(CARD);
    await expect(card.locator('[data-brand-mark]')).toHaveAttribute('data-empty', 'neutral');
    await expect(card.locator('.brand-signin-title')).toHaveText('کنسول مدیریت');
    await card.locator('[data-signin-tools] [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(card.locator('.brand-signin-title')).toHaveText('Admin console');
    await expect(card.locator('.brand-signin-sub')).toHaveText('Sign in to continue');
    await expect(card.locator('[data-signin-tools] [data-locale="en"]')).toHaveAttribute('aria-checked', 'true');
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    const [c, tl] = [await card.boundingBox(), await card.locator('[data-signin-tools]').boundingBox()];
    expect(tl!.x).toBeGreaterThanOrEqual(c!.x);
    expect(tl!.x + tl!.width).toBeLessThanOrEqual(c!.x + c!.width);
    expect(tl!.y + tl!.height).toBeLessThanOrEqual(c!.y + c!.height);
  });
});
