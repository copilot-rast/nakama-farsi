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
