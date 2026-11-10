import { expect, test } from '../../lib/fixtures';
import { base } from '../../lib/api';
import { open } from '../../lib/console';
import type { Page } from '@playwright/test';

async function visit(page: Page, hash: string) {
  await page.goto('/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.waitForFunction(() => !document.querySelector('#main .animate-pulse'));
}
import { restoreConfig } from '../../lib/db';

function ratio(a: number[], b: number[]) {
  const lin = (c: number) => { c /= 255; return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; };
  const lum = (c: number[]) => 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]);
  const x = lum(a), y = lum(b);
  return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05);
}

const rgb = (s: string) => (s.match(/[\d.]+/g) || []).slice(0, 3).map(Number);

test.describe('spatial shell: look', () => {
  test('sky and glass on every main page, swapped by the dark toggle without reload', async ({ page }) => {
    await visit(page, '#/players');
    const html = page.locator('html');
    await expect(html).toHaveClass(/(^|\s)sp(\s|$)/);
    await expect(page.locator('body > .sp-sky svg')).toHaveCount(1);
    await expect(page.locator('.sp-sky')).toHaveAttribute('data-theme', 'light');
    const before = await page.evaluate(() => ({ bg: getComputedStyle(document.documentElement).backgroundColor, body: getComputedStyle(document.body).backgroundColor, main: getComputedStyle(document.getElementById('main')!).backgroundColor }));
    expect(before.body).toBe('rgba(0, 0, 0, 0)');
    expect(before.main).toBe('rgba(0, 0, 0, 0)');
    const marker = await page.evaluate(() => ((window as any).__spMarker = Math.random()));
    await page.evaluate(() => (window as any).NakamaConsole.toggleTheme());
    await expect(html).toHaveClass(/dark/);
    await expect(page.locator('.sp-sky')).toHaveAttribute('data-theme', 'dark');
    await expect(page.locator('.sp-sky circle[fill="#fff"]')).toHaveCount(140);
    expect(await page.evaluate(() => (window as any).__spMarker)).toBe(marker);
    const ink = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--ink-1').trim());
    expect(ink).toBe('#f1f2f6');
    await page.evaluate(() => (window as any).NakamaConsole.toggleTheme());
    await expect(page.locator('.sp-sky')).toHaveAttribute('data-theme', 'light');
  });

  test('?spatial=0 brings back the stock console and is remembered; ?spatial=1 turns it on again', async ({ page }) => {
    await page.goto('/?spatial=0#/players');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor();
    await expect(page.locator('html')).not.toHaveClass(/(^|\s)sp(\s|$)/);
    await expect(page.locator('.sp-sky')).toHaveCount(0);
    await expect(page.locator('.x-drawer').first()).toBeVisible();
    await page.goto('/#/groups');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor();
    await expect(page.locator('html')).not.toHaveClass(/(^|\s)sp(\s|$)/);
    await page.goto('/?spatial=1#/groups');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached' });
    await expect(page.locator('html')).toHaveClass(/(^|\s)sp(\s|$)/);
    expect(await page.evaluate(() => localStorage.getItem('nakama.spatial'))).toBeNull();
  });

  test('the brand accent drives the accent tokens and primary buttons, picked for contrast', async ({ spawn }) => {
    try {
      await base.put('/brand', { name: {}, shortName: '', logoUrl: '', logoDarkUrl: '', markUrl: '', accent: '#059669', welcome: {} });
      const page = await spawn({ signedIn: true });
      await open(page, '#/economy/products');
      await expect.poll(() => page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--sp-accent').trim())).toBe('rgb(5 135 95)');
      const t = await page.evaluate(() => {
        const cs = getComputedStyle(document.documentElement);
        const btn = document.querySelector('#economy-root .x-button[style*="--x-button-bg: rgb(var(--primary-color"]') as HTMLElement;
        return { accent: cs.getPropertyValue('--sp-accent').trim(), ink: cs.getPropertyValue('--sp-accent-ink').trim(), inkDark: cs.getPropertyValue('--sp-accent-ink-dark').trim(), btn: btn ? getComputedStyle(btn).backgroundColor : '', btnText: btn ? getComputedStyle(btn).color : '' };
      });
      expect(t.accent).toBe('rgb(5 135 95)');
      expect(rgb(t.btn)).toEqual([5, 135, 95]);
      expect(ratio(rgb(t.btnText), rgb(t.btn))).toBeGreaterThanOrEqual(4.5);
      expect(ratio(rgb(t.ink), [223, 228, 238])).toBeGreaterThanOrEqual(4.5);
      expect(ratio(rgb(t.inkDark), [74, 77, 95])).toBeGreaterThanOrEqual(4.5);
      await base.put('/brand', { name: {}, shortName: '', logoUrl: '', logoDarkUrl: '', markUrl: '', accent: '#ffee00', welcome: {} });
      const page2 = await spawn({ signedIn: true });
      await open(page2, '#/economy/products');
      await expect.poll(() => page2.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--primary-color-500').trim())).toBe('255 238 0');
      const y = await page2.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--sp-accent').trim());
      expect(ratio([255, 255, 255], rgb(y))).toBeGreaterThanOrEqual(4.5);
    } finally {
      restoreConfig();
    }
  });
});

const MAIN: [string, string][] = [
  ['#/', 'داشبورد'], ['#/players', 'بازیکنان'], ['#/groups', 'گروه‌ها'], ['#/storage', 'ذخیره‌سازی'], ['#/storage-indexes', 'نمایه‌های ذخیره‌سازی'],
  ['#/matches', 'مسابقه‌ها'], ['#/leaderboards', 'جدول‌های امتیازات'], ['#/notifications', 'اعلان‌ها'], ['#/payments', 'پرداخت‌ها'], ['#/chat', 'پیام‌های گفتگو'],
];
const GROUPS: [string, string[]][] = [
  ['economy', ['#/economy/overview', '#/economy/currencies', '#/economy/metas', '#/economy/prizes', '#/economy/products', '#/economy/gifts', '#/economy/payments']],
  ['publication', ['#/publication/faqs', '#/publication/news', '#/publication/comments', '#/publication/notifications']],
  ['developer', ['#/api-explorer', '#/runtime']],
  ['settings', ['#/settings/users', '#/settings/general', '#/settings/config', '#/settings/data-management', '#/settings/audit-log']],
];

const pathOf = (h: string) => h.split('?')[0].replace(/^#/, '');

test.describe('spatial shell: rail and top bar', () => {
  test('every main page is one click away on the rail and marked current, with its title in the top bar', async ({ page }) => {
    await visit(page, '#/');
    await expect(page.locator('.sp-rail')).toBeVisible();
    await expect(page.locator('div.flex.border-r.z-20:has(> .x-drawer)')).toBeHidden();
    for (const [hash, title] of MAIN) {
      await page.locator(`.sp-rail a.sp-ri[href="${hash}"]`).click();
      await expect.poll(() => page.evaluate(() => location.hash || '#/')).toMatch(new RegExp('^' + hash.replace(/[/#?]/g, '\\$&') + '(\\?|$)'));
      await expect(page.locator(`.sp-rail a.sp-ri[href="${hash}"]`)).toHaveAttribute('aria-current', 'page');
      await expect(page.locator('.sp-rail [aria-current]')).toHaveCount(1);
      await expect(page.locator('.sp-crumbs .sp-crumb-current')).toHaveText(title);
    }
  });

  test('every group page is two clicks away through its flyout; the group is marked and the flyout closes', async ({ page }) => {
    await visit(page, '#/');
    for (const [group, hashes] of GROUPS) {
      for (const hash of hashes) {
        const btn = page.locator(`.sp-rail button[data-sp-group="${group}"]`);
        await btn.click();
        const fly = page.locator('.sp-flyout.open');
        await expect(fly).toBeVisible();
        await expect(btn).toHaveAttribute('aria-expanded', 'true');
        await fly.locator(`a[href^="${hash}"]`).click();
        await expect.poll(() => page.evaluate(() => (location.hash || '#/').split('?')[0])).toContain(pathOf(hash));
        await expect(page.locator('.sp-flyout.open')).toHaveCount(0);
        await expect(btn).toHaveAttribute('aria-current', 'true');
      }
    }
  });

  test('flyout opens on hover and closes when the pointer leaves, on Escape and with the keyboard only', async ({ page }) => {
    await visit(page, '#/players');
    const btn = page.locator('.sp-rail button[data-sp-group="publication"]');
    await btn.hover();
    await expect(page.locator('.sp-flyout.open')).toBeVisible();
    await expect(page.locator('.sp-flyout.open .sp-fly-head b')).toHaveText('انتشار');
    await page.mouse.move(700, 450);
    await expect(page.locator('.sp-flyout.open')).toHaveCount(0);
    await page.locator('.sp-rail a.sp-ri[href="#/chat"]').focus();
    await page.keyboard.press('ArrowDown');
    await expect(page.locator('.sp-rail button[data-sp-group="economy"]')).toBeFocused();
    await page.keyboard.press('Enter');
    await expect(page.locator('.sp-flyout.open')).toBeVisible();
    await expect(page.locator('.sp-flyout.open a[role="menuitem"]').first()).toBeFocused();
    await page.keyboard.press('ArrowDown');
    await expect(page.locator('.sp-flyout.open a[role="menuitem"]').nth(1)).toBeFocused();
    await page.keyboard.press('Escape');
    await expect(page.locator('.sp-flyout.open')).toHaveCount(0);
    await expect(page.locator('.sp-rail button[data-sp-group="economy"]')).toBeFocused();
    await page.keyboard.press('Enter');
    await page.keyboard.press('ArrowDown');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/currencies/);
    await expect(page.locator('.sp-rail button[data-sp-group="economy"]')).toHaveAttribute('aria-current', 'true');
  });

  test('the pinned rail survives a reload, lists the sections and collapses back', async ({ page }) => {
    await visit(page, '#/groups');
    await page.locator('.sp-rail .sp-pin').click();
    await expect(page.locator('html')).toHaveClass(/sp-pinned/);
    await expect(page.locator('.sp-rail-full')).toBeVisible();
    const pad = await page.evaluate(() => getComputedStyle(document.getElementById('main')!).paddingInlineStart);
    expect(pad).toBe('264px');
    await page.reload();
    await page.locator('.sp-rail-full a.sp-row[href="#/groups"]').waitFor();
    await expect(page.locator('.sp-rail-full a.sp-row[href="#/groups"]')).toHaveAttribute('aria-current', 'page');
    await page.locator('.sp-rail-full button[data-sp-group="economy"]').click();
    await expect(page.locator('.sp-flyout.open')).toBeVisible();
    await page.locator('.sp-flyout.open a[href^="#/economy/gifts"]').click();
    await expect(page).toHaveURL(/#\/economy\/gifts/);
    await page.locator('.sp-rail-full .sp-collapse').click();
    await expect(page.locator('html')).not.toHaveClass(/sp-pinned/);
    expect(await page.evaluate(() => localStorage.getItem('nakama.spatial.rail'))).toBeNull();
  });

  test('breadcrumbs follow the route; section tabs slide between economy pages; badges match the old sidebar', async ({ page }) => {
    await visit(page, '#/players');
    await page.locator('#main table tbody tr').first().click();
    await expect(page).toHaveURL(/#\/players\/[0-9a-f-]+/);
    await expect(page.locator('.sp-crumbs a.sp-crumb-link')).toHaveText('بازیکنان');
    await expect(page.locator('.sp-crumbs .sp-crumb-current')).toHaveText(/[0-9a-f-]{36}/);
    await page.goto('/#/economy/products');
    await page.locator('.sp-tabs a').first().waitFor();
    await expect(page.locator('.sp-crumbs')).toContainText('اقتصاد');
    await expect(page.locator('.sp-crumbs .sp-crumb-current')).toHaveText('فروشگاه');
    await expect(page.locator('.sp-tabs a')).toHaveCount(7);
    await expect(page.locator('.sp-tabs a[aria-current="page"]')).toHaveText('فروشگاه');
    const x1 = await page.locator('.sp-tabs .sp-seg-ind').evaluate((e) => getComputedStyle(e).transform);
    await page.locator('.sp-tabs a', { hasText: 'ارزها' }).click();
    await expect(page.locator('.sp-tabs a[aria-current="page"]')).toHaveText('ارزها');
    await expect.poll(() => page.locator('.sp-tabs .sp-seg-ind').evaluate((e) => getComputedStyle(e).transform)).not.toBe(x1);
    const old = await page.locator('.x-drawer a[data-economy-page="comments"] .eco-nav-badge').count();
    if (old) {
      const n = await page.locator('.x-drawer a[data-economy-page="comments"] .eco-nav-badge').textContent();
      await expect(page.locator('.sp-rail button[data-sp-group="publication"] .sp-count')).toHaveText(n!.trim());
    }
  });

  test('EN/FA and theme from the top bar; LTR puts the rail on the left', async ({ page }) => {
    await visit(page, '#/players');
    const dark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
    await page.locator('.sp-theme').click();
    await expect.poll(() => page.evaluate(() => document.documentElement.classList.contains('dark'))).toBe(!dark);
    await page.locator('.sp-theme').click();
    await page.locator('.sp-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await page.locator('.sp-rail').waitFor();
    const box = await page.locator('.sp-rail').boundingBox();
    expect(box!.x).toBeLessThan(40);
    await expect(page.locator('.sp-crumbs .sp-crumb-current')).toHaveText('Players');
    await page.locator('.sp-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('project settings opens from the rail and the top bar gear, and closes back to the page', async ({ page }) => {
    await visit(page, '#/groups');
    await page.locator('.sp-rail button[data-sp-group="settings"]').click();
    await page.locator('.sp-flyout.open [data-ps-entry="rail"]').click();
    await expect(page.locator('[data-economy="project-settings"] .ps-window')).toBeVisible();
    await expect(page.locator('.sp-rail button[data-sp-group="settings"]')).toHaveAttribute('aria-current', 'true');
    await page.keyboard.press('Escape');
    await expect(page.locator('[data-economy="project-settings"]')).toHaveCount(0);
    await expect(page).toHaveURL(/#\/groups$/);
    await page.locator('[data-ps-entry="topbar"]').click();
    await expect(page.locator('[data-economy="project-settings"] .ps-window')).toBeVisible();
    await page.locator('[data-ps-close]').click();
    await expect(page).toHaveURL(/#\/groups$/);
  });

  test('search jumps to any page, including project settings sections', async ({ page }) => {
    await visit(page, '#/');
    await page.keyboard.press('Control+k');
    const input = page.locator('.sp-pal-input');
    await expect(input).toBeFocused();
    await input.fill('هدیه');
    await expect(page.locator('.sp-pal-row').first()).toContainText('هدیه‌ها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/gifts/);
    await page.locator('.sp-search').click();
    await page.locator('.sp-pal-input').fill('پیامک');
    await page.keyboard.press('Enter');
    await expect(page.locator('[data-economy="project-settings"] .ps-window')).toBeVisible();
    await expect(page).toHaveURL(/#\/project-settings\/sms/);
  });

  test('sign-in has no rail, a centered glass window on the sky', async ({ spawn }) => {
    const page = await spawn({ signedIn: false });
    await page.goto('/#/login');
    await page.locator('input[name="username"]').waitFor();
    await expect(page.locator('html')).toHaveClass(/sp-auth/);
    await expect(page.locator('.sp-rail')).toHaveCount(0);
    await expect(page.locator('.sp-sky')).toHaveCount(1);
    const box = await page.locator('form.x-form').boundingBox();
    expect(Math.abs(box!.x + box!.width / 2 - 720)).toBeLessThan(40);
  });
});
