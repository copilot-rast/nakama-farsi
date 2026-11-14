import type { Page, Route } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { restoreConfig } from '../../lib/db';
import { noHorizontalScroll, open, root } from '../../lib/console';

const W = '[data-economy="setup"]';

async function summaryWith(page: Page, patch: { fn: (s: any) => void }) {
  await page.route('**/v2/console/economy/summary', async (route: Route) => {
    const res = await route.fetch();
    const s = await res.json();
    patch.fn(s);
    await route.fulfill({ response: res, json: s });
  });
}

async function waitStep(page: Page, step: string) {
  await expect(page.locator(`${W} [data-setup-step="${step}"]`)).toBeVisible();
  await expect(page.locator(`${W} .animate-pulse`)).toHaveCount(0);
}

async function leave(page: Page) {
  const ask = page.locator('[data-economy="dialog"] [role="alertdialog"]');
  if (await ask.isVisible().catch(() => false)) await ask.locator('button').last().click();
}

async function bodyFits(page: Page) {
  return page.locator(`${W} .setup-body`).evaluate((b) => b.scrollHeight - b.clientHeight);
}

test.describe('economy overview', () => {
  test('KPI row, catalog grouped like the sidebar, tiles open their pages, labels stay on one line', async ({ page }) => {
    await open(page, '#/economy/overview');
    const r = root(page);
    const kpis = r.locator('[data-ov-kpis] > [data-ov-kpi]');
    await expect(kpis).toHaveCount(4);
    await expect(r.locator('[data-ov-kpi="revenue"] .eco-ov-spark')).toHaveCount(1);
    await expect(r.locator('[data-ov-kpi="revenue"] [data-ov-delta]')).toHaveCount(1);
    await expect(r.locator('[data-ov-events]')).toContainText('رویدادهای ۲۴ ساعت اخیر');
    await expect(r.locator('[data-ov-group] h3')).toHaveText(['ارز و پیشرفت', 'فروش', 'جایزه و بازی', 'رویداد و رشد']);
    for (const p of ['offers', 'friend-gifts', 'meta-relations', 'items', 'achievements', 'chance-games', 'referrals']) {
      await expect(r.locator(`[data-ov-tile="${p}"]`)).toHaveCount(1);
    }
    const wraps = await r.locator('.eco-ov-tile-label, .eco-ov-tile-sub').evaluateAll((els) =>
      els.filter((e) => e.getBoundingClientRect().height > parseFloat(getComputedStyle(e).lineHeight) * 1.5).map((e) => e.textContent));
    expect(wraps).toEqual([]);
    const holes = await r.locator('[data-ov-group] .eco-ov-tiles').evaluateAll((grids) => grids.map((g) => {
      const kids = Array.from(g.children) as HTMLElement[];
      const rows = new Map<number, number>();
      kids.forEach((k) => { const top = Math.round(k.getBoundingClientRect().top); rows.set(top, (rows.get(top) || 0) + k.getBoundingClientRect().width); });
      const full = g.getBoundingClientRect().width;
      return Array.from(rows.values()).filter((w) => w < full - 40).length;
    }));
    expect(holes.every((n) => n === 0)).toBe(true);

    for (const [tile, url] of [['offers', /#\/economy\/offers$/], ['event-rules', /#\/economy\/events\/prizes$/], ['tags', /#\/project-settings\/tags$/]] as const) {
      await r.locator(`[data-ov-tile="${tile}"]`).click();
      await expect(page).toHaveURL(url);
      await page.goBack();
      await expect(r.locator('[data-ov-kpis]')).toBeVisible();
    }
  });

  test('payments: 30 daily bars, metric switch, status rows link to the filtered list', async ({ page }) => {
    await open(page, '#/economy/overview');
    const r = root(page);
    const pay = r.locator('[data-ov-payments]');
    await expect(pay.locator('[data-ov-chart="revenue"] .eco-ov-bar')).toHaveCount(30);
    await pay.locator('[data-ov-metric] [data-value="payments"]').click();
    await expect(pay.locator('[data-ov-chart="payments"]')).toBeVisible();
    await expect(pay.locator('[data-ov-chart="payments"] svg')).toHaveAttribute('aria-label', /پرداخت/);
    await expect(pay.locator('[data-ov-status-row]')).toHaveCount(4);
    await pay.locator('[data-ov-status-row="failed"]').click();
    await expect(page).toHaveURL(/#\/economy\/payments\?status=failed$/);
  });

  test('empty states: no active gateway links to payment settings, no payments links to the shop', async ({ page }) => {
    const patch = { fn: (s: any) => {
      s.paymentMethods = s.paymentMethods.map((m: any) => ({ ...m, isActive: false }));
      s.counts.payments = 0;
      s.paymentsByStatus = { pending: 0, success: 0, failed: 0, expired: 0 };
      s.revenueDaily = s.revenueDaily.map((d: any) => ({ ...d, toman: 0, payments: 0, success: 0 }));
    } };
    await summaryWith(page, patch);
    await open(page, '#/economy/overview');
    const r = root(page);
    await expect(r.locator('[data-ov-empty-methods]')).toContainText('هیچ درگاه پرداختی فعال نیست');
    await expect(r.locator('.eco-ov-plot-empty')).toContainText('درآمد واقعی ثبت نشده');
    await expect(r.locator('[data-ov-status]')).toHaveCount(0);
    await r.locator('[data-ov-cta="data-ov-empty-methods"]').click();
    await expect(page).toHaveURL(/#\/project-settings\/payment-methods$/);

    patch.fn = (s: any) => {
      s.paymentMethods = s.paymentMethods.map((m: any, i: number) => ({ ...m, isActive: i === 0 }));
      s.counts.payments = 0;
    };
    await open(page, '#/economy/overview');
    await page.reload();
    await root(page).waitFor();
    await expect(r.locator('[data-ov-kpis]')).toBeVisible();
    await expect(r.locator('[data-ov-empty-methods]')).toHaveCount(0);
    await expect(r.locator('[data-ov-empty-payments] [data-ov-cta]')).toHaveAttribute('href', '#/economy/products');
  });

  test('English and 390px without sideways scroll', async ({ page }) => {
    await open(page, '#/economy/overview');
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(root(page).locator('[data-ov-group="sales"] h3')).toHaveText('Sales');
    await expect(root(page).locator('[data-ov-kpi="revenue"]')).toContainText('Revenue, last 30 days');
    await page.setViewportSize({ width: 390, height: 844 });
    await noHorizontalScroll(page);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
  });
});

test.describe('setup wizard stepper', () => {
  test.afterEach(() => restoreConfig());

  test('one step at a time, fits 1280×720 without scrolling, progress, back and next', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await expect(page.locator(W)).toBeVisible();
    await expect(page.locator(`${W} .setup-step`)).toHaveCount(7);
    for (const [i, step] of ['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review'].entries()) {
      await page.locator(`${W} .setup-step[data-step="${step}"]`).click();
      await leave(page);
      await waitStep(page, step);
      await expect(page.locator(`${W} [data-setup-of]`)).toHaveText(`مرحلهٔ ${'۱۲۳۴۵۶۷'[i]} از ۷`);
      if (i > 0) await expect(page.locator(`${W} #setup-step-title`)).toBeFocused();
      if (step === 'sms' || step === 'email') {
        const yes = page.locator(`${W} [data-need] [data-value="yes"]`);
        if ((await yes.getAttribute('aria-checked')) !== 'true') await yes.click();
      }
      expect(await bodyFits(page), step + ' scrolls').toBeLessThanOrEqual(1);
    }
    await page.locator(`${W} [data-setup-back]`).click();
    await leave(page);
    await waitStep(page, 'email');
    await expect(page.locator(`${W} .setup-step[data-step="email"]`)).toHaveAttribute('aria-current', 'step');
    await expect(page.locator(`${W} [role="progressbar"]`)).toHaveAttribute('aria-valuemax', '6');
  });

  test('extras fold under more settings; summary lists every step with an edit link', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await page.locator(`${W} .setup-step[data-step="brand"]`).click();
    await leave(page);
    await waitStep(page, 'brand');
    const more = page.locator(`${W} [data-setup-more]`);
    await expect(more).not.toHaveAttribute('open', '');
    await expect(page.locator(`${W} input[aria-label="نام کوتاه"]`)).toBeHidden();
    await more.locator('summary').focus();
    await page.keyboard.press('Enter');
    await expect(page.locator(`${W} input[aria-label="نام کوتاه"]`)).toBeVisible();
    await expect(page.locator(`${W} .brand-preview`)).toBeVisible();
    const btn = page.locator(`${W} .brand-preview-btn`).first();
    const swatch = (await btn.evaluate((e) => getComputedStyle(e).backgroundColor)) === 'rgb(5, 135, 95)' ? '#2563eb' : '#059669';
    const before = await btn.evaluate((e) => getComputedStyle(e).backgroundColor);
    await page.locator(`${W} .brand-swatch[data-color="${swatch}"]`).click();
    await expect.poll(() => btn.evaluate((e) => getComputedStyle(e).backgroundColor)).not.toBe(before);

    await page.locator(`${W} .setup-step[data-step="review"]`).click();
    await leave(page);
    await waitStep(page, 'review');
    await expect(page.locator(`${W} [data-review]`)).toHaveCount(6);
    await page.locator(`${W} [data-review="currency"] button`).click();
    await waitStep(page, 'currency');
    await expect(page.locator(`${W} input[aria-label="موجودی اول"]`)).toBeVisible();
    await page.keyboard.press('Escape');
    const ask = page.locator('[data-economy="dialog"] [role="alertdialog"]');
    await expect(ask).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator(W)).toBeVisible();
  });

  test('phone: the step bar is a compact row and only the step body scrolls', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await page.locator(`${W} .setup-step[data-step="payment"]`).click();
    await leave(page);
    await waitStep(page, 'payment');
    const win = await page.locator(`${W} .setup-window`).boundingBox();
    expect(win!.width).toBeLessThanOrEqual(390);
    await expect(page.locator(`${W} [data-setup-next]`)).toBeInViewport();
    await expect(page.locator(`${W} .setup-step[data-step="payment"] .setup-step-label`)).toBeVisible();
    await expect(page.locator(`${W} .setup-step[data-step="brand"] .setup-step-label`)).toBeHidden();
    await noHorizontalScroll(page);
  });
});
