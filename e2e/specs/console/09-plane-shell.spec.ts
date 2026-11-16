import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';

async function visit(page: Page, hash: string) {
  await page.goto('/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.locator('.pl-side').waitFor();
  await page.waitForFunction(() => !document.querySelector('#main .animate-pulse'));
}

const MODES: [string, string, string][] = [
  ['game', 'بازی', '#/'],
  ['economy', 'اقتصاد', '#/economy/overview'],
  ['publication', 'انتشار', '#/publication/faqs'],
  ['moderation', 'بازیکنان و گفتگو', '#/moderation/players'],
  ['support', 'پشتیبانی', '#/support/tickets'],
  ['system', 'سامانه', '#/system/updates'],
];

test.describe('plane shell: look', () => {
  test('flat canvas, white sidebar, Yekan Bakh, and the dark toggle swaps tokens without a reload', async ({ page }) => {
    await visit(page, '#/players');
    await expect(page.locator('html')).toHaveClass(/\bpl\b/);
    await expect(page.locator('.pl-rail')).toBeVisible();
    await expect(page.locator('.pl-side')).toBeVisible();
    const font = await page.evaluate(() => getComputedStyle(document.body).fontFamily);
    expect(font).toContain('Yekan Bakh');
    await page.evaluate(() => document.fonts.ready);
    expect(await page.evaluate(() => document.fonts.check('14px "Yekan Bakh"'))).toBe(true);
    const shadow = await page.locator('.pl-side').evaluate((e) => getComputedStyle(e).boxShadow);
    expect(shadow).toBe('none');
    const light = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--base').trim());
    await page.locator('button.pl-theme').click();
    await expect(page.locator('html')).toHaveClass(/\bdark\b/);
    const dark = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--base').trim());
    expect(dark).not.toBe(light);
    await page.locator('button.pl-theme').click();
    await expect(page.locator('html')).not.toHaveClass(/\bdark\b/);
  });

  test('?plane=0 brings back the stock console and is remembered; ?plane=1 turns it on again', async ({ page }) => {
    await page.goto('/?plane=0#/players');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached' });
    await expect(page.locator('html')).not.toHaveClass(/\bpl\b/);
    await expect(page.locator('.pl-rail')).toHaveCount(0);
    await page.goto('/#/players');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached' });
    await expect(page.locator('html')).not.toHaveClass(/\bpl\b/);
    await page.goto('/?plane=1#/players');
    await page.locator('.pl-rail').waitFor();
    await expect(page.locator('html')).toHaveClass(/\bpl\b/);
  });

  test('the dashboard has icon tiles on every KPI and a needs-attention card', async ({ page }) => {
    await visit(page, '#/');
    await expect(page.locator('#main .pl-kpi .pl-kpi-ic svg.pl-duo').first()).toBeVisible();
    expect(await page.locator('#main .pl-kpi .pl-kpi-ic').count()).toBeGreaterThanOrEqual(4);
    await expect(page.locator('#main .pl-attn h2')).toHaveText('نیاز به رسیدگی');
  });
});

test.describe('plane shell: rail, sidebar and top bar', () => {
  test('each rail mode shows its pages in the sidebar and is the only current mode', async ({ page }) => {
    await visit(page, '#/players');
    for (const [id, label, hash] of MODES) {
      const mode = page.locator(`.pl-rail .pl-mode[data-pl-mode="${id}"]`);
      await expect(mode).toHaveAttribute('aria-label', label);
      await mode.click();
      await expect(page).toHaveURL(new RegExp(hash.replace(/[/#?]/g, '\\$&')));
      await expect(page.locator('.pl-side .pl-side-label').first()).toHaveText(label);
      await expect(page.locator('.pl-rail [aria-current]')).toHaveCount(1);
      await expect(mode).toHaveAttribute('aria-current', 'true');
      await expect(page.locator('.pl-side a.pl-nav[aria-current="page"]')).toHaveCount(1);
      const icons = await page.locator('.pl-side a.pl-nav').evaluateAll((as) => as.every((a) => !!a.querySelector('svg')));
      expect(icons).toBe(true);
    }
  });

  test('economy sidebar lists its 18 entries with their icons, current state and breadcrumbs', async ({ page }) => {
    await visit(page, '#/economy/products');
    const side = page.locator('.pl-side a.pl-nav');
    await expect(side).toHaveCount(18);
    for (const [hash, label] of [['#/economy/tasks', 'تسک‌ها'], ['#/economy/items', 'آیتم‌ها'], ['#/economy/achievements', 'دستاوردها'], ['#/economy/chance-games', 'بازی شانس'], ['#/economy/prizes', 'جایزه‌ها'], ['#/economy/games', 'بازی‌ها'], ['#/economy/events', 'رویدادها']]) {
      const link = page.locator(`.pl-side a.pl-nav[href="${hash}"]`);
      await expect(link).toContainText(label);
      await expect(link.locator('svg')).toHaveCount(1);
      await link.click();
      await expect(link).toHaveAttribute('aria-current', 'page');
      await expect(page.locator('.pl-crumbs .pl-crumb-current')).toHaveText(label);
      await expect(page.locator('#economy-root h1, #economy-root h2').first()).toContainText(label);
    }
  });

  test('breadcrumbs follow the route and sidebar badges match the old sidebar', async ({ page }) => {
    await visit(page, '#/economy/currencies');
    await expect(page.locator('.pl-crumbs .pl-crumb-current')).toHaveText('ارزها');
    await expect(page.locator('.pl-crumbs')).toContainText('اقتصاد');
    await visit(page, '#/moderation/players');
    const old = await page.locator('.x-drawer a[href^="#/moderation/reports"] .eco-nav-badge').first().textContent({ timeout: 2000 }).catch(() => null);
    const badge = page.locator('.pl-side a[href^="#/moderation/reports"] .pl-badge');
    if (old && old.trim()) await expect(badge).toHaveText(old.trim());
    else await expect(badge).toHaveCount(0);
  });

  test('EN/FA from the top bar; LTR puts the rail on the left', async ({ page }) => {
    await visit(page, '#/economy/currencies');
    let r = await page.locator('.pl-rail').boundingBox();
    expect(r!.x).toBeGreaterThan(1000);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(page.locator('.pl-rail .pl-mode[data-pl-mode="economy"]')).toHaveAttribute('aria-label', 'Economy');
    r = await page.locator('.pl-rail').boundingBox();
    expect(r!.x).toBe(0);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('project settings opens from the rail settings view and the top bar gear, and closes back to the page', async ({ page }) => {
    await visit(page, '#/publication/news');
    await page.locator('.pl-rail .pl-mode[data-pl-mode="settings"]').click();
    await page.locator('.pl-side [data-ps-entry="rail"]').click();
    await expect(page.locator('.ps-window')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('.ps-window')).toHaveCount(0);
    await expect(page).toHaveURL(/#\/publication\/news$/);
    await page.locator('[data-ps-entry="topbar"]').click();
    await expect(page.locator('.ps-window')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page).toHaveURL(/#\/publication\/news$/);
  });

  test('search from the sidebar jumps to any page, including project settings sections', async ({ page }) => {
    await visit(page, '#/players');
    await page.locator('.pl-side-search').click();
    await page.locator('.pl-pal-input').fill('هدیه');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/gifts/);
    await expect(page.locator('.pl-pal-input')).toHaveCount(0);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('پیامک');
    await expect(page.locator('.pl-pal-row').first()).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('.pl-pal-input')).toHaveCount(0);
  });

  test('sign-in has no rail and no sidebar', async ({ spawn }) => {
    const page = await spawn({ signedIn: false });
    await page.goto('/#/login');
    await page.locator('input[name="username"]').waitFor();
    await expect(page.locator('html')).toHaveClass(/pl-auth/);
    await expect(page.locator('.pl-rail:visible')).toHaveCount(0);
    await expect(page.locator('.pl-side:visible')).toHaveCount(0);
  });
});

test.describe('plane switch', () => {
  test('every switch shows a white knob inside its track, on the end side when on and the start side when off', async ({ page }) => {
    for (const hash of ['#/project-settings/languages', '#/players']) {
      await visit(page, hash);
      const toggles = page.locator('.x-toggle:visible');
      await toggles.first().waitFor();
      const states = await toggles.evaluateAll((ts) => ts.map((t) => {
        const track = t.querySelector('.rounded-full.border-\\[3px\\]')!;
        const knob = track.querySelector('.relative > .rounded-full')!;
        const a = track.getBoundingClientRect();
        const k = knob.getBoundingClientRect();
        const on = (t.querySelector('input') as HTMLInputElement).getAttribute('aria-checked') === 'true';
        const rtl = getComputedStyle(t).direction === 'rtl';
        const mid = k.left + k.width / 2 - a.left;
        const endSide = rtl ? mid < a.width / 2 : mid > a.width / 2;
        return { inside: k.left >= a.left - 0.5 && k.right <= a.right + 0.5, visible: k.width >= 14 && getComputedStyle(knob).backgroundColor === 'rgb(255, 255, 255)', ok: on === endSide, hit: t.getBoundingClientRect().height };
      }));
      expect(states.length).toBeGreaterThan(0);
      for (const st of states) {
        expect(st.inside).toBe(true);
        expect(st.visible).toBe(true);
        expect(st.ok).toBe(true);
      }
    }
  });
});

test.describe('plane motion', () => {
  test('no transform on #main or the app root while an editor is open; transitions stay short', async ({ page }) => {
    await visit(page, '#/economy/products/new');
    await page.locator('[data-economy="panel"] .x-form').waitFor();
    const bad = await page.evaluate(() => ['#main', '#app', 'body'].filter((s) => { const n = document.querySelector(s); return n && getComputedStyle(n).transform !== 'none'; }));
    expect(bad).toEqual([]);
    const d = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--d3').trim());
    expect(parseInt(d, 10)).toBeLessThanOrEqual(300);
  });
});
