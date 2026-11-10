import { expect, test } from '../../lib/fixtures';
import { API_URL, CONSOLE_URL, SERVER_KEY } from '../../lib/env';
import { base, consoleApi, deleteUser, devicePlayer, must, pub, rpc, uid, updateAccount } from '../../lib/api';
import { codeCount, otpSandboxOn, removePhones, resetOtpLimits, testPhone, waitCode } from '../../lib/accounts';
import { restoreConfig } from '../../lib/db';
import { field, fillTranslated, go, loginAs, open, openSettings, panel, psWindow, root, rows, savePanel, search, settle, settlePanel, toasts, translated } from '../../lib/console';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  await removePhones();
  restoreConfig();
});

test.afterEach(() => restoreConfig());

function form(page: import('@playwright/test').Page, id: string) {
  return root(page).locator(`[data-settings-form="${id}"]`);
}

async function saveForm(page: import('@playwright/test').Page, id: string) {
  await form(page, id).locator('[data-settings-save]').click();
  await expect(toasts(page).filter({ hasText: 'ذخیره شد' }).last()).toBeVisible();
}

test.describe('accounts: countries', () => {
  test('seeded list with flags, search, inline toggle, filter, create, edit and delete a country', async ({ page }) => {
    await open(page, '#/project-settings/countries');
    await expect(psWindow(page).locator('#ps-section-title')).toHaveText('کشورها');
    await expect(rows(page).first()).toContainText('ایران');
    await expect(rows(page).first()).toContainText('+98');
    await expect(root(page).locator('.eco-flag').first()).toHaveText('🇮🇷');

    await search(page, 'آلمان');
    const de = rows(page).filter({ hasText: 'آلمان' });
    await expect(de).toHaveCount(1);
    await expect(de).toContainText('+49');
    await de.locator('.x-toggle').click();
    await expect(toasts(page).filter({ hasText: 'آلمان غیرفعال شد' })).toBeVisible();
    await expect(de).toContainText('غیرفعال');
    expect((await base.get('/countries/DE')).isActive).toBe(false);
    const player = await devicePlayer('e2e-ct-' + k);
    players.push(player.id);
    const list = await rpc(player, 'base_countries', { lang: 'fa' });
    expect(list.data.items.some((c: any) => c.code === 'DE')).toBe(false);
    expect(list.data.items[0]).toMatchObject({ code: 'IR', name: 'ایران', dialCode: '+98', flag: '🇮🇷' });

    await go(page, '#/project-settings/countries?isActive=false');
    await expect(root(page).locator('.eco-filters select').first()).toHaveValue('false');
    await expect(rows(page).filter({ hasText: 'آلمان' })).toHaveCount(1);

    await go(page, '#/project-settings/countries/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کد کشور$/).locator('input').fill('Q1');
    await field(p, /^پیش‌شماره$/).locator('input').fill('+0');
    await fillTranslated(p, 'name', { fa: 'کشور آزمایشی ' + k, en: 'Test land ' + k });
    await p.getByRole('button', { name: 'ایجاد' }).click();
    await expect(field(p, /^کد کشور$/)).toContainText('دو حرف لاتین');
    await expect(field(p, /^پیش‌شماره$/)).toContainText('پیش‌شماره با +');
    await field(p, /^کد کشور$/).locator('input').fill('qm');
    await field(p, /^پیش‌شماره$/).locator('input').fill('999');
    await savePanel(page);
    const qm = await base.get('/countries/QM');
    expect(qm).toMatchObject({ code: 'QM', dialCode: '+999', isActive: true, name: { fa: 'کشور آزمایشی ' + k, en: 'Test land ' + k } });

    await go(page, '#/project-settings/countries?q=' + encodeURIComponent(k));
    await expect(rows(page)).toHaveCount(1);
    await rows(page).first().click();
    await settlePanel(page);
    await expect(field(panel(page), /^کد کشور$/).locator('input')).toHaveAttribute('readonly', '');
    await fillTranslated(panel(page), 'name', { fa: 'کشور آزمایشی دوم ' + k });
    await savePanel(page);
    await expect(rows(page).first()).toContainText('کشور آزمایشی دوم ' + k);
    await rows(page).first().click();
    await settlePanel(page);
    await panel(page).getByRole('button', { name: 'حذف' }).click();
    await panel(page).getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(panel(page)).toHaveCount(0);
    expect((await consoleApi('GET', '/v2/console/base/countries/QM')).status).toBe(404);
  });
});

test.describe('accounts: rules', () => {
  test('rules form validates, the try-it box follows the numbers, the server enforces the saved rules', async ({ page }) => {
    await openSettings(page, 'accounts');
    const f = form(page, 'accounts');
    await expect(f.locator('select')).toHaveValue('any');
    await field(f, /^کمترین طول$/).locator('input').fill('9');
    await field(f, /^بیشترین طول$/).locator('input').fill('6');
    await f.locator('[data-settings-save]').click();
    await expect(field(f, /^بیشترین طول$/)).toContainText('بیشترین طول نمی‌تواند از کمترین طول کمتر باشد.');
    await field(f, /^کمترین طول$/).locator('input').fill('5');
    await field(f, /^بیشترین طول$/).locator('input').fill('12');
    await field(f, /^حروف مجاز$/).locator('select').selectOption('latin');
    await field(f, /^کمترین طول رمز$/).locator('input').fill('10');
    await f.locator('.eco-field', { hasText: 'رمز عدد داشته باشد' }).locator('.x-toggle').click();

    await f.locator('[data-rules-try="username"]').fill('abc');
    await expect(f.locator('[data-rules-bad]')).toHaveText(['نام کاربری دست‌کم ۵ حرف است.']);
    await f.locator('[data-rules-try="username"]').fill('سارا_جان');
    await expect(f.locator('[data-rules-bad]')).toHaveText(['نام کاربری فقط حرف لاتین، رقم و _ . - دارد.']);
    await f.locator('[data-rules-try="username"]').fill('sara_99');
    await f.locator('[data-rules-try="password"]').fill('abcdefghij');
    await expect(f.locator('[data-rules-bad]')).toHaveText(['رمز باید عدد داشته باشد.']);
    await f.locator('[data-rules-try="password"]').fill('abcdefghi1');
    await expect(f.locator('[data-rules-ok]')).toHaveText('پذیرفته می‌شود.');
    await saveForm(page, 'accounts');
    expect(await base.get('/account-rules')).toMatchObject({ usernameMin: 5, usernameMax: 12, usernameCharset: 'latin', passwordMin: 10, passwordDigit: true, passwordLetter: false });

    await page.reload();
    await settle(page);
    await expect(field(form(page, 'accounts'), /^کمترین طول$/).locator('input')).toHaveValue('۵');

    const player = await devicePlayer('e2e-rules-' + k);
    players.push(player.id);
    const tooShort = await fetch(API_URL + '/v2/account', { method: 'PUT', headers: { Authorization: 'Bearer ' + player.token, 'Content-Type': 'application/json' }, body: JSON.stringify({ username: 'abc' }) });
    expect(tooShort.status).toBe(400);
    expect((await tooShort.json()).message).toBe('USERNAME_TOO_SHORT');
    const persian = await fetch(API_URL + '/v2/account', { method: 'PUT', headers: { Authorization: 'Bearer ' + player.token, 'Content-Type': 'application/json' }, body: JSON.stringify({ username: 'سارا_جان' }) });
    expect((await persian.json()).message).toBe('USERNAME_INVALID_CHARACTERS');
    await updateAccount(player, { username: 'e2e' + k.slice(-8) });

    const email = 'e2e-' + k + '@example.com';
    const signUp = (password: string) => fetch(API_URL + '/v2/account/authenticate/email?create=true', {
      method: 'POST',
      headers: { Authorization: 'Basic ' + Buffer.from(SERVER_KEY + ':').toString('base64'), 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    const weak = await signUp('abcdefghij');
    expect(weak.status).toBe(400);
    expect((await weak.json()).message).toBe('PASSWORD_NEEDS_DIGIT');
    const ok = await signUp('abcdefghi1');
    expect(ok.status).toBe(200);
    const claims = JSON.parse(Buffer.from((await ok.json()).token.split('.')[1], 'base64url').toString('utf8'));
    players.push(claims.uid);
    expect((await rpc(player, 'base_account_rules')).data).toMatchObject({ usernameMin: 5, usernameMax: 12, usernameCharset: 'latin', passwordMin: 10 });
  });
});

test.describe('accounts: SMS login', () => {
  test('settings show the sandbox mode, validate the SMS text, and recent requests never show the code', async ({ page }) => {
    expect(otpSandboxOn(), 'SMS login must be in sandbox before the e2e run').toBe(true);
    resetOtpLimits();
    await openSettings(page, 'otp');
    await expect(root(page).locator('.eco-notice--warning')).toContainText('حالت آزمایشی');
    const f = form(page, 'otp');
    await expect(field(f, /^سرویس پیامک$/).locator('select')).toHaveValue('');
    await field(f, /^تعداد رقم کد$/).locator('input').fill('9');
    await f.locator('[data-settings-save]').click();
    await expect(field(f, /^تعداد رقم کد$/)).toHaveClass(/eco-invalid/);
    await field(f, /^تعداد رقم کد$/).locator('input').fill('6');
    await field(f, /^کد در ساعت برای هر شماره$/).locator('input').fill('8');
    await field(f, /^کد در روز برای هر شماره$/).locator('input').fill('4');
    await f.locator('[data-settings-save]').click();
    await expect(field(f, /^کد در روز برای هر شماره$/)).toContainText('سقف روزانه نمی‌تواند از سقف ساعتی کمتر باشد.');
    await field(f, /^کد در روز برای هر شماره$/).locator('input').fill('20');
    await fillTranslated(f, 'message', { fa: 'رمز بازی ' + k });
    await f.locator('[data-settings-save]').click();
    await expect(field(f, /^متن پیامک$/)).toContainText('متن هر زبان باید {code} داشته باشد.');
    await fillTranslated(f, 'message', { fa: 'کد ستاره‌چین: {code}', en: 'Star Catcher code: {code}' });
    await saveForm(page, 'otp');
    const saved = await pub.get('/otp/settings');
    expect(saved).toMatchObject({ codeLength: 6, phoneHourly: 8, phoneDaily: 20, sandbox: true, isEnabled: true, message: { fa: 'کد ستاره‌چین: {code}', en: 'Star Catcher code: {code}' } });

    const player = await devicePlayer('e2e-otp-' + k);
    players.push(player.id);
    const phone = testPhone();
    const before = codeCount(phone.e164);
    const req = await rpc(player, 'publication_otp_request', { phone: phone.local });
    expect(req.status).toBe(200);
    expect(req.data).toMatchObject({ phone: phone.e164, codeLength: 6, sandbox: true });
    const code = await waitCode(phone.e164, before);
    expect(code).toMatch(/^\d{6}$/);
    const again = await rpc(player, 'publication_otp_request', { phone: phone.local });
    expect(again.data.message || again.data).toContain('OTP_TOO_SOON');

    await go(page, '#/project-settings/otp');
    await search(page, phone.e164.slice(3));
    const row = rows(page).filter({ hasText: phone.e164 });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText('منتظر کد');
    await expect(row).toContainText('آزمایشی');
    expect(await root(page).innerText()).not.toContain(code);

    const verify = await rpc(player, 'publication_otp_verify', { phone: phone.e164, code, link: true });
    expect(verify.status).toBe(200);
    expect(verify.data).toMatchObject({ linked: true, userId: player.id, phone: phone.e164 });
    await go(page, '#/project-settings/otp?status=verified&q=' + phone.e164.slice(3));
    await expect(rows(page).filter({ hasText: phone.e164 })).toContainText('وارد شد');
    await expect(rows(page).filter({ hasText: phone.e164 }).locator(`a[href="#/players/${player.id}"]`)).toBeVisible();
  });
});

test.describe('accounts: app versions', () => {
  test('create a platform row with store links and messages; the client RPC answers ok, update or forced update', async ({ page }) => {
    await open(page, '#/project-settings/app-versions');
    await expect(rows(page).filter({ hasText: 'android' })).toContainText('کافه‌بازار');
    await root(page).getByRole('button', { name: 'نسخه‌ی تازه' }).click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^پلتفرم$/).locator('select').selectOption('ios');
    await field(p, /^آخرین نسخه$/).locator('input').fill('2.0.0');
    await field(p, /^کمترین نسخه$/).locator('input').fill('2.1');
    await p.getByRole('button', { name: 'افزودن لینک فروشگاه' }).click();
    const link = p.locator('[data-store-row]').first();
    await link.locator('select').selectOption('app_store');
    await link.locator('[data-link-url]').fill('ftp://bad');
    await p.getByRole('button', { name: 'ایجاد' }).click();
    await expect(field(p, /^کمترین نسخه$/)).toContainText('کمترین نسخه نمی‌تواند از آخرین نسخه تازه‌تر باشد.');
    await expect(link).toContainText('لینک https');
    await expect(link.locator('.eco-items-err')).not.toHaveText('');
    await field(p, /^کمترین نسخه$/).locator('input').fill('1.5.0');
    await link.locator('[data-link-url]').fill('https://apps.apple.com/app/id' + k);
    await fillTranslated(p, 'updateMessage', { fa: 'نسخه‌ی تازه رسید ' + k, en: 'New version ' + k });
    await fillTranslated(p, 'forceMessage', { fa: 'حتماً به‌روز کن ' + k, en: 'You must update ' + k });
    await savePanel(page);
    const row = rows(page).filter({ hasText: 'ios' });
    await expect(row).toContainText('2.0.0');
    await expect(row).toContainText('1.5.0');
    await expect(row).toContainText('اپ‌استور');

    const player = await devicePlayer('e2e-ver-' + k);
    players.push(player.id);
    const ask = async (version: string, lang = 'fa') => (await rpc(player, 'base_app_version', { platform: 'ios', version, lang })).data;
    expect(await ask('1.0', 'en')).toMatchObject({ status: 'update_required', message: 'You must update ' + k, storeUrl: 'https://apps.apple.com/app/id' + k, minimum: '1.5.0' });
    expect(await ask('1.9.3')).toMatchObject({ status: 'update_available', message: 'نسخه‌ی تازه رسید ' + k, latest: '2.0.0' });
    expect((await ask('1.9.3')).stores).toEqual([{ store: 'app_store', name: 'اپ‌استور', url: 'https://apps.apple.com/app/id' + k }]);
    expect(await ask('2.0.0')).toMatchObject({ status: 'ok', message: '' });

    await row.click();
    await settlePanel(page);
    await panel(page).locator('.eco-field', { hasText: 'خاموش باشد، بازی همیشه' }).locator('.x-toggle').click();
    await savePanel(page);
    await expect(rows(page).filter({ hasText: 'ios' })).toContainText('غیرفعال');
    expect(await ask('1.0')).toMatchObject({ status: 'ok' });
  });
});

test.describe('accounts: access', () => {
  test('a console user with notification access sees SMS login but not the settings-only sections', async ({ spawn }) => {
    const user = { user: 'e2epo' + k.slice(4, 12), pass: 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!' };
    const add = await consoleApi('POST', '/v2/console/user', { username: user.user, email: user.user + '@example.com', acl: { ACCOUNT: { read: true }, NOTIFICATION: { read: true } } });
    expect(add.status).toBeLessThan(300);
    try {
      const reset = await must(consoleApi('POST', '/v2/console/user/' + user.user + '/reset/password'));
      const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: reset.code || add.data?.token, password: user.pass }) });
      expect(pc.ok).toBe(true);
      const p = await spawn({ signedIn: false });
      await loginAs(p, user);
      await p.goto(CONSOLE_URL + '/#/project-settings/otp');
      await expect(psWindow(p)).toBeVisible();
      const ids = await psWindow(p).locator('.ps-list [role="tab"]').evaluateAll((els) => els.map((e) => e.getAttribute('data-ps-section')));
      expect(ids).toEqual(['otp', 'sms', 'email']);
      await settle(p);
      await expect(root(p).locator('[data-settings-form="otp"] form')).toBeVisible();
      await p.goto(CONSOLE_URL + '/#/project-settings/countries');
      await expect(p).toHaveURL(/#\/project-settings\/otp$/);
    } finally {
      await consoleApi('DELETE', '/v2/console/user?username=' + user.user);
    }
  });
});
