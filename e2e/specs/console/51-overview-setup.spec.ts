import type { Page, Route } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { restoreConfig } from '../../lib/db';
import { noHorizontalScroll, open, root } from '../../lib/console';
import { base, eco, pub } from '../../lib/api';

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
      await expect(page.locator(`${W} [data-setup-progress]`)).toContainText('مرحله انجام شده');
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

  test('phone: the step bar is a compact row and only the step body scrolls', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, mode: 'dark', width: 390, height: 844, mobile: true });
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

  test('full run in Persian: every step saves through the API, required fields are checked, reopening shows the saved values', async ({ page: first, spawn }) => {
    let page = first;
    const tag = 'e2e' + Date.now().toString(36);
    await page.setViewportSize({ width: 1280, height: 800 });
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await page.locator(`${W} .setup-step[data-step="brand"]`).click();
    await leave(page);
    await waitStep(page, 'brand');
    await expect(page.locator(`${W} [data-setup-more] > summary`)).not.toContainText('فاوآیکن');
    const name = page.locator(`${W} [data-translated="name"] [role="tabpanel"][data-lang="fa"] input`);
    await name.fill('برند ' + tag);
    await page.keyboard.press('Control+Enter');
    await waitStep(page, 'languages');
    expect((await base.get('/brand')).name.fa).toBe('برند ' + tag);

    const known = (await base.get('/languages/known')).items.map((l: any) => l.code);
    const have = (await base.get('/languages?limit=100')).items.map((l: any) => l.code);
    await expect(page.locator(`${W} [data-add-lang]`)).toHaveCount(known.filter((c: string) => !have.includes(c)).length);
    await expect(page.locator(`${W} [data-setup-next]`)).toHaveText('ادامه');
    await page.locator(`${W} [data-setup-next]`).click();
    await waitStep(page, 'currency');

    await page.locator(`${W} input[aria-label="موجودی اول"]`).fill('۷۷');
    await page.locator(`${W} [data-setup-more] > summary`).click();
    await expect(page.locator(`${W} select[aria-label="نوع ارز"], ${W} [aria-label="نوع ارز"]`).first()).toBeVisible();
    await page.locator(`${W} [data-setup-next]`).click();
    await waitStep(page, 'payment');
    const main = (await eco.get('/currencies?limit=100')).items.sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0))[0];
    expect(main.defaultValue).toBe(77);

    const yes = page.locator(`${W} [data-need] [data-value="yes"]`);
    if ((await yes.getAttribute('aria-checked')) !== 'true') await yes.click();
    const bazaar = page.locator(`${W} [data-method="bazaar"]`);
    await expect(bazaar.locator('.setup-method-desc')).toContainText('کافه‌بازار');
    if (!(await bazaar.evaluate((el) => el.classList.contains('setup-method--on')))) await bazaar.locator('[data-pick-method="bazaar"]').click();
    if ((await bazaar.locator('[data-method-edit]').getAttribute('aria-expanded')) !== 'true') await bazaar.locator('[data-method-edit]').click();
    const pkg = bazaar.locator('input[data-field="packageName"]');
    await pkg.fill('');
    await page.locator(`${W} [data-setup-next]`).click();
    await expect(page.locator(`${W} [data-setup-errors]`)).toBeVisible();
    await expect(pkg).toBeFocused();
    await expect(pkg).toHaveAttribute('aria-invalid', 'true');
    await expect(bazaar.locator('.eco-invalid')).toContainText('این فیلد لازم است.');
    await expect(page.locator(`${W} [data-setup-step="payment"]`)).toBeVisible();
    await pkg.fill('ir.' + tag + '.game');
    await expect(pkg).not.toHaveAttribute('aria-invalid', 'true');
    await page.locator(`${W} [data-setup-next]`).click();
    await waitStep(page, 'sms');
    expect((await eco.get('/payment-methods')).find((m: any) => m.method === 'bazaar').config.packageName).toBe('ir.' + tag + '.game');

    const smsActive = (await pub.get('/providers')).items.some((p: any) => p.channel === 'sms' && p.isActive);
    await page.locator(`${W} [data-need] [data-value="no"]`).click();
    await page.locator(`${W} [data-setup-next]`).click();
    await waitStep(page, 'email');
    const email = page.locator(`${W} [data-need] [data-value="yes"]`);
    if ((await email.getAttribute('aria-checked')) !== 'true') await email.click();
    if ((await page.locator(`${W} [data-kind] [aria-checked="true"]`).getAttribute('data-value')) !== 'smtp') await page.locator(`${W} [data-kind] [data-value="smtp"]`).click();
    await expect(page.locator(`${W} [data-sandbox-box]`)).toContainText('ایمیلی واقعاً فرستاده نمی‌شود');
    await expect(page.locator(`${W} [data-sandbox-box]`)).not.toContainText('پرداخت');
    await expect(page.locator(`${W} [data-test-otp]`)).toHaveCount(0);
    const host = page.locator(`${W} input[aria-label="میزبان"]`);
    await host.fill('');
    await page.locator(`${W} [data-setup-next]`).click();
    await expect(page.locator(`${W} [data-setup-errors]`)).toBeVisible();
    await expect(host).toHaveAttribute('aria-invalid', 'true');
    await host.fill('smtp.' + tag + '.test');
    const from = page.locator(`${W} input[aria-label="فرستنده"]`);
    if (!(await from.inputValue())) await from.fill('noreply@example.com');
    await page.locator(`${W} [data-setup-next]`).click();
    await waitStep(page, 'review');
    const smtp = (await pub.get('/providers')).items.find((p: any) => p.kind === 'smtp');
    expect(smtp).toMatchObject({ isActive: true, isDefault: true });
    expect(smtp.config.host).toBe('smtp.' + tag + '.test');
    await expect(page.locator(`${W} [data-review="brand"]`)).toContainText('برند ' + tag);
    await expect(page.locator(`${W} [data-review="currency"]`)).toContainText('۷۷');
    await page.locator(`${W} [data-setup-next]`).click();
    await expect(page.locator(W)).toHaveCount(0);
    expect((await base.get('/setup')).status).toBe('completed');

    page = await spawn({ signedIn: true, width: 1280, height: 800 });
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await page.locator(`${W} .setup-step[data-step="brand"]`).click();
    await waitStep(page, 'brand');
    await expect(page.locator(`${W} [data-translated="name"] [role="tabpanel"][data-lang="fa"] input`)).toHaveValue('برند ' + tag);
    await page.locator(`${W} .setup-step[data-step="currency"]`).click();
    await waitStep(page, 'currency');
    await expect(page.locator(`${W} input[aria-label="موجودی اول"]`)).toHaveValue(/^(77|۷۷)$/);
    await page.locator(`${W} .setup-step[data-step="payment"]`).click();
    await waitStep(page, 'payment');
    await expect(page.locator(`${W} [data-need] [data-value="yes"]`)).toHaveAttribute('aria-checked', 'true');
    await page.locator(`${W} [data-method="bazaar"] [data-method-edit]`).click();
    await expect(page.locator(`${W} [data-method="bazaar"] input[data-field="packageName"]`)).toHaveValue('ir.' + tag + '.game');
    await page.locator(`${W} .setup-step[data-step="sms"]`).click();
    await waitStep(page, 'sms');
    await expect(page.locator(`${W} [data-need] [data-value="${smsActive ? 'yes' : 'no'}"]`)).toHaveAttribute('aria-checked', 'true');
    await page.locator(`${W} .setup-step[data-step="email"]`).click();
    await waitStep(page, 'email');
    await expect(page.locator(`${W} input[aria-label="میزبان"]`)).toHaveValue('smtp.' + tag + '.test');
  });

  test('English and dark: keyboard through the step bar, Ctrl+Enter, English labels and errors, no sideways scroll at 390px', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, locale: 'en', mode: 'dark', width: 1280, height: 800 });
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await page.locator(`${W} .setup-step[data-step="brand"]`).click();
    await leave(page);
    await waitStep(page, 'brand');
    await expect(page.locator(`${W} [data-setup-of]`)).toHaveText('Step 1 of 7');
    await expect(page.locator(`${W} [data-setup-progress]`)).toContainText('steps done');
    await page.locator(`${W} .setup-step[data-step="brand"]`).focus();
    await page.keyboard.press('ArrowDown');
    await expect(page.locator(`${W} .setup-step[data-step="languages"]`)).toBeFocused();
    await page.keyboard.press('End');
    await expect(page.locator(`${W} .setup-step[data-step="review"]`)).toBeFocused();
    await page.keyboard.press('Home');
    await expect(page.locator(`${W} .setup-step[data-step="brand"]`)).toBeFocused();
    await page.keyboard.press('ArrowDown');
    await page.keyboard.press('Enter');
    await leave(page);
    await waitStep(page, 'languages');
    await expect(page.locator(`${W} [data-setup-next]`)).toHaveText('Continue');
    await page.keyboard.press('Control+Enter');
    await waitStep(page, 'currency');
    await expect(page.locator(`${W} [data-setup-step="currency"] h2`)).toHaveText('Main game currency');
    await page.locator(`${W} .setup-step[data-step="payment"]`).click();
    await leave(page);
    await waitStep(page, 'payment');
    const yes = page.locator(`${W} [data-need] [data-value="yes"]`);
    if ((await yes.getAttribute('aria-checked')) !== 'true') await yes.click();
    const myket = page.locator(`${W} [data-method="myket"]`);
    if (!(await myket.evaluate((el) => el.classList.contains('setup-method--on')))) await myket.locator('[data-pick-method="myket"]').click();
    if ((await myket.locator('[data-method-edit]').getAttribute('aria-expanded')) !== 'true') await myket.locator('[data-method-edit]').click();
    await myket.locator('input[data-field="packageName"]').fill('');
    await page.keyboard.press('Control+Enter');
    await expect(myket.locator('.eco-invalid')).toContainText('This field is required.');
    await expect(page.locator(`${W} [data-setup-errors]`)).toContainText('missing');
    expect(await page.evaluate(() => getComputedStyle(document.querySelector('.setup-window')!).backgroundColor)).not.toBe('rgb(255, 255, 255)');
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator(`${W} [data-setup-next]`)).toBeInViewport();
    await noHorizontalScroll(page);
    await page.keyboard.press('Escape');
    await page.locator('[data-economy="dialog"] [role="alertdialog"] button').last().click();
  });
});
