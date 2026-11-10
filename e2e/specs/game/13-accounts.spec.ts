import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { base, deleteUser, uid } from '../../lib/api';
import { codeCount, otpSandboxOn, removePhones, resetOtpLimits, testPhone, waitCode } from '../../lib/accounts';
import { restoreConfig } from '../../lib/db';
import { GAME_URL } from '../../lib/env';
import { openGame, persian, player, screen, toast, waitHome } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.beforeAll(() => {
  expect(otpSandboxOn(), 'SMS login must be in sandbox before the e2e run').toBe(true);
  resetOtpLimits();
});

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  await removePhones();
  restoreConfig();
});

async function start(page: Page, tag: string, query?: string) {
  const p = await openGame(page, `e2e-acc-${tag}-${k}`, query ? { query } : {});
  players.push(p.id);
  return p;
}

async function toAccount(page: Page) {
  await page.locator('#menu-account').click();
  await screen(page, 'account');
}

async function signIn(page: Page, phone: { local: string; e164: string }, opts: { keep?: boolean; wrongFirst?: boolean } = {}) {
  await page.locator('#otp-phone').fill(phone.local);
  if (opts.keep === false) await page.locator('#otp-link').uncheck();
  resetOtpLimits();
  const before = codeCount(phone.e164);
  await page.locator('#otp-send').click();
  await expect(page.locator('#otp-step-code')).toBeVisible();
  const code = await waitCode(phone.e164, before);
  if (opts.wrongFirst) {
    await page.locator('#otp-code').fill(code === '00000' ? '11111' : '00000');
    await page.locator('#otp-verify').click();
    await expect(page.locator('#otp-msg')).toHaveText('کد درست نیست.');
  }
  await page.locator('#otp-code').fill(code);
  await page.locator('#otp-verify').click();
  await expect(page.locator('#btn-signout')).toBeVisible();
}

test.describe('stargame: phone sign-in', () => {
  test('a guest links a phone, keeps the account over a reload and reaches it from a second device', async ({ page, spawn }) => {
    const a = await start(page, 'a');
    await toAccount(page);
    await expect(page.locator('#account-title')).toHaveText('مهمان');
    const phone = testPhone();
    await page.locator('#otp-phone').fill(phone.local);
    const before = codeCount(phone.e164);
    await page.locator('#otp-send').click();
    await expect(page.locator('#otp-step-code')).toBeVisible();
    await expect(page.locator('#otp-sent')).toContainText('حالت آزمایشی');
    await expect(page.locator('#otp-resend')).toBeDisabled();
    await expect(page.locator('#otp-resend')).toContainText('ارسال دوباره تا');
    await expect(page.locator('#otp-code')).toHaveAttribute('maxlength', '5');
    const code = await waitCode(phone.e164, before);
    await page.locator('#otp-code').fill(code === '00000' ? '11111' : '00000');
    await page.locator('#otp-verify').click();
    await expect(page.locator('#otp-msg')).toHaveText('کد درست نیست.');
    await page.locator('#otp-code').fill(persian(code));
    await page.locator('#otp-verify').click();
    await expect(toast(page)).toContainText('به همین حساب وصل شد');
    await expect(page.locator('#account-title')).toContainText(phone.e164);
    await expect(page.locator('#otp-form')).toBeHidden();
    expect((await player(page)).id).toBe(a.id);
    expect(await page.evaluate(() => localStorage.getItem('stargame.loginKey'))).toMatch(/^ph_[0-9a-f]{40}$/);

    await page.reload();
    await waitHome(page);
    expect((await player(page)).id).toBe(a.id);
    await toAccount(page);
    await expect(page.locator('#account-title')).toContainText(phone.e164);

    const second = await spawn({ signedIn: false });
    const b = await openGame(second, `e2e-acc-b-${k}`);
    players.push(b.id);
    expect(b.id).not.toBe(a.id);
    await toAccount(second);
    await signIn(second, phone);
    await expect(toast(second)).toContainText('خوش آمدی');
    expect((await player(second)).id).toBe(a.id);
    await second.locator('#btn-signout').click();
    await expect(toast(second)).toContainText('از حساب موبایل خارج شدی');
    await expect(second.locator('#account-title')).toHaveText('مهمان');
    expect((await player(second)).id).toBe(b.id);
  });

  test('a new phone account when progress is not kept; the country picker saves my country and hides disabled ones', async ({ page }) => {
    const c = await start(page, 'c');
    await toAccount(page);
    await page.locator('#btn-country').click();
    await page.locator('#country-q').fill('آلمان');
    await expect(page.locator('#country-list [data-code]')).toHaveCount(1);
    await page.locator('#country-list [data-code="DE"]').click();
    await expect(toast(page)).toContainText('کشورت شد آلمان');
    await expect(page.locator('#btn-country')).toContainText('🇩🇪');
    await expect(page.locator('#otp-country')).toContainText('+49');
    await page.locator('#otp-country').click();
    await page.locator('#country-q').fill('98');
    await page.locator('#country-list [data-code="IR"]').click();
    await expect(page.locator('#otp-country')).toContainText('+98');

    await page.reload();
    await waitHome(page);
    await toAccount(page);
    await expect(page.locator('#btn-country')).toContainText('آلمان');
    await expect(page.locator('#otp-country')).toContainText('+49');
    await page.locator('#otp-country').click();
    await page.locator('#country-q').fill('ایران');
    await page.locator('#country-list [data-code="IR"]').click();
    await expect(page.locator('#otp-country')).toContainText('+98');

    const phone = testPhone();
    await signIn(page, phone, { keep: false });
    await expect(toast(page)).toContainText('حساب تازه');
    const now = await player(page);
    expect(now.id).not.toBe(c.id);
    players.push(now.id);

    await base.put('/countries/DE', { ...(await base.get('/countries/DE')), isActive: false });
    await page.reload();
    await waitHome(page);
    await toAccount(page);
    await page.locator('#btn-country').click();
    await page.locator('#country-q').fill('آلمان');
    await expect(page.locator('#country-list')).toContainText('کشوری پیدا نشد.');
    await page.keyboard.press('Escape');
    restoreConfig();
  });

  test('username rules from the console show in the game and the server enforces them', async ({ page }) => {
    await base.put('/account-rules', { usernameMin: 6, usernameMax: 14, usernameCharset: 'latin', passwordMin: 8, displayNameMax: 32 });
    try {
      await start(page, 'n');
      await toAccount(page);
      await expect(page.locator('#name-hint')).toHaveText('بین ۶ تا ۱۴ حرف، فقط حرف لاتین، رقم و _ . -.');
      await page.locator('#name-input').fill('abc');
      await page.locator('#name-form button[type="submit"]').click();
      await expect(page.locator('#name-msg')).toHaveText('نام کاربری کوتاه است.');
      await page.locator('#name-input').fill('ستاره_چین');
      await page.locator('#name-form button[type="submit"]').click();
      await expect(page.locator('#name-msg')).toHaveText('نام کاربری فقط حرف لاتین، رقم و _ . - می‌تواند داشته باشد.');
      const name = 'e2e' + k.slice(-9);
      await page.locator('#name-input').fill(name);
      await page.locator('#name-form button[type="submit"]').click();
      await expect(page.locator('#name-msg')).toHaveText('نام کاربری ذخیره شد.');
      await page.locator('#screen-account [data-go="home"]').click();
      await expect(page.locator('#username')).toHaveText(name);
    } finally {
      restoreConfig();
    }
  });
});

test.describe('stargame: version check', () => {
  test('an optional update shows a banner, a forced one blocks the game; both follow the console', async ({ page, spawn }) => {
    await start(page, 'v1', '?v=1.3.0');
    await waitHome(page);
    const banner = page.locator('#update-banner');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText(/نسخه(‌ی|ٔ) تازه(‌ی|ٔ) ستاره‌چین آمده/);
    await expect(page.locator('#update-banner-go')).toHaveAttribute('href', /\/stargame\/$/);
    await expect(page.locator('#game-version')).toHaveText('نسخهٔ 1.3.0');
    await page.locator('#update-banner-later').click();
    await expect(banner).toBeHidden();
    await page.reload();
    await waitHome(page);
    await expect(banner).toBeHidden();

    const old = await spawn({ signedIn: false });
    const o = await openGame(old, `e2e-acc-v2-${k}`, { query: '?v=0.9.0' });
    players.push(o.id);
    await screen(old, 'update');
    await expect(old.locator('#screen-update h1')).toHaveText('باید بازی را به‌روز کنی');
    await expect(old.locator('#update-text')).toHaveText('این نسخه دیگر کار نمی‌کند. برای ادامه، بازی را به‌روز کن.');
    await expect(old.locator('#update-versions')).toContainText('0.9.0');
    await expect(old.locator('#update-stores a[data-store="web"]')).toHaveText('وب‌سایت');
    await old.locator('.topbar .brand').click();
    await screen(old, 'update');
    await expect(old.locator('.topbar-end')).toBeHidden();

    const web = (await base.get('/app-versions?q=web')).items.find((x: any) => x.platform === 'web');
    try {
      await base.put('/app-versions/' + web.id, { ...web, latestVersion: '1.6.0', minVersion: '1.5.0', forceMessage: { fa: 'نسخه‌ی ' + k + ' لازم است', en: 'Version ' + k + ' needed' } });
      const now = await spawn({ signedIn: false });
      const n = await openGame(now, `e2e-acc-v3-${k}`, { query: '?lang=fa' });
      players.push(n.id);
      await screen(now, 'update');
      await expect(now.locator('#update-text')).toHaveText('نسخه‌ی ' + k + ' لازم است');
    } finally {
      restoreConfig();
    }
    const fresh = await spawn({ signedIn: false });
    const f = await openGame(fresh, `e2e-acc-v4-${k}`);
    players.push(f.id);
    await expect(fresh.locator('#update-banner')).toBeHidden();
    await expect(fresh.locator('#game-version')).toHaveText('نسخهٔ 1.4.0');
  });
});
