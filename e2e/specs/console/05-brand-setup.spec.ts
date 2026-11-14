import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, devicePlayer, deleteUser, eco, must, rpc, uid } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { ADMIN, loginAs, open, root, settle, toasts } from '../../lib/console';
import { CONSOLE_URL, FIXTURES } from '../../lib/env';
import type { Page } from '@playwright/test';

const W = '[data-economy="setup"]';
const k = uid();

test.describe.configure({ mode: 'serial' });
test.afterEach(() => restoreConfig());

async function waitStep(page: Page, step: string) {
  await expect(page.locator(`${W} [data-setup-step="${step}"]`)).toBeVisible();
  await expect(page.locator(`${W} .animate-pulse`)).toHaveCount(0);
}

async function next(page: Page) {
  await page.locator(`${W} [data-setup-next]`).click();
}

async function later(page: Page) {
  await page.locator(`${W} [data-setup-later]`).click();
  const ask = page.locator('[data-economy="dialog"] [role="alertdialog"]');
  if (await ask.isVisible().catch(() => false)) await ask.locator('button').last().click();
  await expect(page.locator(W)).toHaveCount(0);
}

async function applied(page: Page) {
  return page.evaluate(() => ({
    title: document.title,
    favicon: document.querySelector('link[rel~="icon"]')?.getAttribute('href') || '',
    p500: getComputedStyle(document.documentElement).getPropertyValue('--primary-color-500').trim(),
  }));
}

test.describe('brand and setup wizard', () => {
  test('first login opens the wizard once; brand, languages, currency, payment, later, resume, finish', async ({ spawn }) => {
    await base.put('/setup', { status: 'pending' });
    const page = await spawn({ signedIn: false });
    await loginAs(page, ADMIN);
    await expect(page.locator(W)).toBeVisible({ timeout: 10_000 });
    await waitStep(page, 'brand');
    await expect(page.locator(`${W} #setup-step-title`)).toHaveText('برند پروژه');

    const name = page.locator(`${W} [data-translated="name"]`);
    await name.locator('[role="tabpanel"][data-lang="fa"] input').fill('ستاره‌چین ' + k);
    await name.locator('[role="tab"][data-lang="en"]').click();
    await name.locator('[role="tabpanel"][data-lang="en"] input').fill('Starpicker ' + k);
    await page.locator(`${W} [data-setup-more] > summary`).click();
    await page.locator(`${W} input[aria-label="نام کوتاه"]`).fill('ستاره');
    const files = page.locator(`${W} input[type="file"]`);
    for (const [i, f] of ['logo.svg', 'logo-dark.svg', 'mark.svg'].entries()) {
      const up = page.waitForResponse((r) => r.url().includes('/v2/console/economy/uploads'));
      await files.nth(i).setInputFiles(path.join(FIXTURES, f));
      expect((await up).ok()).toBe(true);
    }
    const hex = page.locator(`${W} .brand-hex input`);
    await hex.fill('#ffee00');
    await expect(page.locator('[data-contrast]').first()).toContainText(/کم‌کنتراست|فقط برای متن درشت/);
    await hex.fill('zz');
    await next(page);
    await expect(page.locator(`${W} .eco-invalid`)).toContainText('هگز');
    await page.locator(`${W} .brand-swatch[data-color="#059669"]`).click();
    await next(page);
    await waitStep(page, 'languages');
    await expect.poll(async () => (await applied(page)).p500).toBe('5 150 105');
    const a = await applied(page);
    expect(a.title).toContain('ستاره‌چین ' + k);
    expect(a.favicon).toMatch(/\/v2\/economy\/files\/.+\.svg$/);

    await page.locator(`${W} li[data-lang="ar"] .x-toggle`).click();
    await expect.poll(async () => (await base.get('/languages/ar')).isActive).toBe(false);
    await page.locator(`${W} li[data-lang="ar"] .x-toggle`).click();
    await expect.poll(async () => (await base.get('/languages/ar')).isActive).toBe(true);
    const hasTr = (await base.get('/languages?limit=100')).items.some((l: any) => l.code === 'tr');
    if (!hasTr) {
      await page.locator(`${W} [data-add-lang="tr"]`).click();
      await expect(page.locator(`${W} li[data-lang="tr"]`)).toBeVisible();
    }
    await next(page);
    await waitStep(page, 'currency');
    await page.locator(`${W} input[aria-label="موجودی اول"]`).fill('۱۲۰');
    await next(page);
    await waitStep(page, 'payment');
    const coin = (await eco.get('/currencies?limit=50')).items.find((c: any) => c.key === 'coin');
    expect(coin.defaultValue).toBe(120);

    const zibal = page.locator(`${W} [data-method="zibal"]`);
    if (!(await zibal.evaluate((el) => el.classList.contains('setup-method--on')))) await zibal.locator('[data-pick-method="zibal"]').click();
    if ((await zibal.locator('[data-method-edit]').getAttribute('aria-expanded')) !== 'true') await zibal.locator('[data-method-edit]').click();
    await expect(zibal.locator('[data-sandbox-box]')).toHaveClass(/eco-sandbox-note/);
    await zibal.locator('input[aria-label="کد مرچنت"]').fill('zibal');
    await next(page);
    await waitStep(page, 'sms');
    const z = (await eco.get('/payment-methods')).find((m: any) => m.method === 'zibal');
    expect(z).toMatchObject({ isActive: true, configured: true });
    expect(z.config.sandbox).toBe(true);
    await page.locator(`${W} [data-need] [data-value="no"]`).click();
    await next(page);
    await waitStep(page, 'email');
    let st = await base.get('/setup');
    expect(st.steps).toMatchObject({ brand: 'done', languages: 'done', currency: 'done', payment: 'done', sms: 'skipped' });

    await page.locator(`${W} .setup-step[data-step="brand"]`).click();
    await waitStep(page, 'brand');
    await expect(page.locator(`${W} [data-translated="name"] [role="tabpanel"][data-lang="fa"] input`)).toHaveValue('ستاره‌چین ' + k);
    await page.locator(`${W} .setup-step[data-step="email"]`).click();
    await waitStep(page, 'email');
    await later(page);
    await expect.poll(async () => (await base.get('/setup')).status).toBe('skipped');
    st = await base.get('/setup');
    expect(st.nextStep).toBe('email');
    await expect(page.locator('.x-drawer [data-brand-logo]').first()).toBeAttached();
    await expect(page.locator('.pl-rail .pl-brand img')).toBeVisible();

    await page.reload();
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached' });
    await page.waitForTimeout(1500);
    await expect(page.locator(W)).toHaveCount(0);

    await page.evaluate(() => { location.hash = '#/project-settings/setup'; });
    await expect(page.locator('[data-setup-row="email"]')).toBeVisible();
    await expect(page.locator('[data-setup-row="sms"]')).toContainText('رد شد');
    await expect(page.locator('[data-setup-row="brand"]')).toContainText('انجام شد');
    await page.locator('[data-setup-open]').click();
    await waitStep(page, 'email');
    if (await page.locator(`${W} [data-need] [data-value="yes"]`).isVisible()) await page.locator(`${W} [data-need] [data-value="yes"]`).click();
    const kind = await page.locator(`${W} [data-kind] [aria-checked="true"]`).getAttribute('data-value').catch(() => null);
    if (kind !== 'smtp') await page.locator(`${W} [data-kind] [data-value="smtp"]`).click();
    await expect(page.locator(`${W} [data-sandbox-box]`).first()).toHaveClass(/eco-sandbox-note/);
    await page.locator(`${W} input[aria-label="گیرنده"]`).fill('qa@example.com');
    await page.locator(`${W} button`, { hasText: 'ذخیره و ارسال آزمایشی' }).click();
    await expect(page.locator(`${W} [data-test-result]`)).toHaveAttribute('data-test-result', 'sandbox');
    await next(page);
    await waitStep(page, 'review');
    await expect(page.locator(`${W} [data-review]`).first()).toBeVisible();
    await next(page);
    await expect(page.locator(W)).toHaveCount(0);
    st = await base.get('/setup');
    expect(st.status).toBe('completed');
    expect(st.completedBy).toBe(ADMIN.user);
  });

  test('step bar opens every step, skip marks a step, Esc asks first, focus stays inside', async ({ page }) => {
    await open(page, '#/project-settings/setup');
    await page.locator('[data-setup-open]').click();
    await expect(page.locator(W)).toBeVisible();
    for (const step of ['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review']) {
      await page.locator(`${W} .setup-step[data-step="${step}"]`).click();
      await waitStep(page, step);
      await expect(page.locator(`${W} .setup-step[data-step="${step}"]`)).toHaveAttribute('aria-current', 'step');
    }
    await page.locator(`${W} .setup-step[data-step="currency"]`).click();
    await waitStep(page, 'currency');
    await page.locator(`${W} [data-setup-skip]`).click();
    await waitStep(page, 'payment');
    await expect(page.locator(`${W} .setup-step[data-step="currency"]`)).toHaveClass(/setup-step--skipped/);
    await expect.poll(async () => (await base.get('/setup')).steps.currency).toBe('skipped');
    await page.keyboard.press('Escape');
    const ask = page.locator('[data-economy="dialog"] [role="alertdialog"]');
    await expect(ask).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(ask).toHaveCount(0);
    await expect(page.locator(W)).toBeVisible();
    for (let i = 0; i < 30; i++) {
      await page.keyboard.press('Tab');
      expect(await page.evaluate(() => !!document.activeElement?.closest('[data-economy="setup"]'))).toBe(true);
    }
    await later(page);
    expect((await base.get('/setup')).status).toBe('completed');
  });

  test('brand shows on sign-in, sidebar, title and favicon; reset brings Nakama back', async ({ spawn }) => {
    await base.put('/brand', {
      name: { fa: 'ستاره‌چین ' + k, en: 'Starpicker ' + k },
      shortName: 'ستاره',
      logoUrl: '', logoDarkUrl: '', markUrl: '',
      accent: '#2563eb',
      welcome: { fa: 'به پنل ستاره‌چین خوش آمدید.', en: 'Welcome to the Starpicker panel.' },
    });
    const fa = await spawn({ signedIn: false });
    await fa.goto(CONSOLE_URL + '/#/login');
    await expect(fa.locator('[data-brand-signin]')).toContainText('ستاره‌چین ' + k);
    await expect(fa.locator('[data-brand-signin]')).toContainText('خوش آمدید');
    await expect(fa.locator('[data-brand-signin]')).toContainText('برای ادامه به داشبورد');
    await expect.poll(async () => (await applied(fa)).p500).toBe('37 99 235');
    expect((await applied(fa)).title).toContain('ستاره‌چین ' + k);

    const en = await spawn({ signedIn: false, locale: 'en', mode: 'dark' });
    await en.goto(CONSOLE_URL + '/#/login');
    await expect(en.locator('html')).toHaveClass(/\bdark\b/);
    await expect(en.locator('[data-brand-signin]')).toContainText('Starpicker ' + k);
    await expect(en.locator('[data-brand-signin]')).toContainText('Welcome to the Starpicker panel.');
    await loginAs(en, ADMIN);
    await expect.poll(async () => (await applied(en)).title).toContain('Starpicker ' + k);

    const dev = await devicePlayer('e2e-brand-' + k);
    const r = await rpc(dev, 'base_brand', { lang: 'en' });
    expect(r.data).toMatchObject({ custom: true, name: 'Starpicker ' + k, accent: '#2563eb', welcome: 'Welcome to the Starpicker panel.' });
    await deleteUser(dev.id);

    await open(en, '#/project-settings/brand');
    await expect(root(en).locator('.brand-editor input').first()).toBeVisible();
    await base.put('/brand', {});
    const fresh = await spawn({ signedIn: false });
    await fresh.goto(CONSOLE_URL + '/#/login');
    await fresh.locator('input[name="username"]').waitFor();
    await fresh.waitForTimeout(800);
    await expect(fresh.locator('[data-brand-logo],[data-brand-signin]')).toHaveCount(0);
    await expect(fresh.locator('h1').first()).toContainText('Nakama');
    await loginAs(fresh, ADMIN);
    await expect.poll(async () => (await applied(fresh)).favicon).toBe('/favicon.ico');
    const after = await applied(fresh);
    expect(after.title).toContain('Nakama');
    expect(after.p500).toMatch(/107\s+89\s+238/);
    expect(await fresh.locator('#brand-accent').count()).toBe(0);
  });

  test('claim once: of two admins signing in together exactly one gets the wizard', async ({ spawn }) => {
    const qa = { user: 'e2eqa' + k.slice(4, 12), pass: 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!' };
    const add = await consoleApi('POST', '/v2/console/user', { username: qa.user, email: qa.user + '@example.com', acl: { ACCOUNT: { read: true }, SETTINGS: { read: true, write: true, delete: true }, HIRO_ECONOMY: { read: true, write: true }, NOTIFICATION: { read: true, write: true } } });
    expect(add.status).toBeLessThan(300);
    try {
      const reset = await must(consoleApi('POST', '/v2/console/user/' + qa.user + '/reset/password'));
      const code = reset.code || add.data?.token;
      const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code, password: qa.pass }) });
      expect(pc.ok).toBe(true);

      const second = await spawn({ signedIn: false });
      await loginAs(second, qa);
      await second.waitForTimeout(1500);
      await expect(second.locator(W)).toHaveCount(0);
      await second.evaluate(() => { location.hash = '#/project-settings/setup'; });
      await second.locator('[data-setup-open]').click();
      await expect(second.locator(`${W} [data-setup-step]`)).toBeVisible();
      await later(second);
      expect((await base.get('/setup')).status).toBe('completed');

      await base.put('/setup', { status: 'pending' });
      const a = await spawn({ signedIn: false });
      const b = await spawn({ signedIn: false });
      await Promise.all([loginAs(a, ADMIN), loginAs(b, qa)]);
      await expect.poll(async () => (await a.locator(W).count()) + (await b.locator(W).count()), { timeout: 10_000 }).toBe(1);
      await a.waitForTimeout(1500);
      const openA = await a.locator(W).count();
      const openB = await b.locator(W).count();
      expect(openA + openB).toBe(1);
      const st = await base.get('/setup');
      expect(st.status).toBe('in_progress');
      expect(st.firstOpenedBy).toBe(openA ? ADMIN.user : qa.user);
      await later(openA ? a : b);
      const c = await spawn({ signedIn: false });
      await loginAs(c, ADMIN);
      await c.waitForTimeout(1500);
      await expect(c.locator(W)).toHaveCount(0);
    } finally {
      await consoleApi('DELETE', '/v2/console/user?username=' + qa.user);
    }
  });
});
