import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, root, toasts } from '../../lib/console';
import { codeCount, otpSandboxOn, removePhones, resetOtpLimits, testPhone, waitCode } from '../../lib/accounts';
import { consoleApi, deleteUser, devicePlayer, rpc, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { API_URL, SERVER_KEY } from '../../lib/env';

const made: string[] = [];

async function emailLogin(email: string, password: string) {
  const r = await fetch(API_URL + '/v2/account/authenticate/email?create=false', {
    method: 'POST',
    headers: { Authorization: 'Basic ' + Buffer.from(SERVER_KEY + ':').toString('base64'), 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  return { status: r.status, data: await r.json().catch(() => null) };
}

function userIdOfToken(token: string) {
  return JSON.parse(Buffer.from(token.split('.')[1], 'base64url').toString()).uid as string;
}

test.describe.configure({ mode: 'serial' });

test.describe('console: create a player without verification', () => {
  test.afterAll(async () => {
    for (const id of made) await deleteUser(id);
    await removePhones();
  });

  test('by phone from the players page; the player then signs in to it with an SMS code', async ({ page }) => {
    expect(otpSandboxOn(), 'SMS login must be in sandbox before the e2e run').toBe(true);
    resetOtpLimits();
    const phone = testPhone();
    const name = uid('pc');
    const sentBefore = codeCount(phone.e164);
    await open(page, '#/moderation/players');
    await root(page).locator('[data-player-new]').first().click();
    const form = page.locator('[data-player-form]');
    await expect(form).toBeVisible();
    await expect(form).toContainText('نه ایمیل تأیید فرستاده می‌شود، نه پیامک');
    await form.getByLabel('شمارهٔ موبایل', { exact: true }).fill(phone.local.replace(/\d/g, (d) => '۰۱۲۳۴۵۶۷۸۹'[+d]));
    await form.getByLabel('نام کاربری').fill(name);
    await form.getByLabel('نام نمایشی').fill('بازیکن آزمایشی');
    await page.locator('[data-player-create]').click();
    await expect(toasts(page).filter({ hasText: name })).toBeVisible();
    await expect(page).toHaveURL(/#\/players\/[0-9a-f-]{36}$/);
    const id = page.url().split('/').pop()!;
    made.push(id);
    expect(sql(`SELECT phone FROM publication_phone_account WHERE user_id='${id}'`)).toBe(phone.e164);
    expect(sql(`SELECT username || '|' || display_name || '|' || COALESCE(metadata->>'phone', 'none') FROM users WHERE id='${id}'`)).toBe(name + '|بازیکن آزمایشی|none');
    await expect(page.locator(`[data-player-phone="${phone.e164}"]`)).toBeVisible();
    expect(codeCount(phone.e164)).toBe(sentBefore);

    const guest = await devicePlayer('e2e-pc-' + Date.now());
    const req = await rpc(guest, 'publication_otp_request', { phone: phone.local });
    expect(req.status).toBe(200);
    const code = await waitCode(phone.e164, sentBefore);
    const verify = await rpc(guest, 'publication_otp_verify', { phone: phone.e164, code });
    expect(verify.status).toBe(200);
    expect(verify.data.userId).toBe(id);
    expect(verify.data.created).toBe(false);
    await deleteUser(guest.id);

    const dup = await consoleApi('POST', '/v2/console/publication/players', { kind: 'phone', phone: phone.e164 });
    expect(dup.status).toBe(409);
    expect(JSON.stringify(dup.data)).toContain('PHONE_TAKEN');
  });

  test('by email: no confirmation email, the player signs in with the password at once', async ({ page }) => {
    const email = uid('pc') + '@example.com';
    await open(page, '#/moderation/players');
    await root(page).locator('[data-player-new]').first().click();
    const form = page.locator('[data-player-form]');
    await form.locator('select').first().selectOption('email');
    await form.getByLabel('ایمیل', { exact: true }).fill(email);
    await form.getByLabel('رمز', { exact: true }).fill('short');
    await page.locator('[data-player-create]').click();
    await expect(form).toContainText('رمز باید ۸ تا ۱۲۸ حرف باشد.');
    await form.getByLabel('رمز', { exact: true }).fill('e2e-pass-1234');
    await page.locator('[data-player-create]').click();
    await expect(page).toHaveURL(/#\/players\/[0-9a-f-]{36}$/);
    const id = page.url().split('/').pop()!;
    made.push(id);
    expect(sql(`SELECT email || '|' || (metadata->>'emailVerified') FROM users WHERE id='${id}'`)).toBe(email + '|true');
    const login = await emailLogin(email, 'e2e-pass-1234');
    expect(login.status).toBe(200);
    expect(userIdOfToken(login.data.token)).toBe(id);
    expect((await emailLogin(email, 'wrong-pass-9999')).status).toBe(401);

    await open(page, '#/moderation/players');
    await root(page).locator('[data-player-new]').first().click();
    const again = page.locator('[data-player-form]');
    await again.locator('select').first().selectOption('email');
    await again.getByLabel('ایمیل', { exact: true }).fill(email.toUpperCase());
    await again.getByLabel('رمز', { exact: true }).fill('e2e-pass-1234');
    await page.locator('[data-player-create]').click();
    await expect(again).toContainText('این ایمیل مال بازیکن دیگری است.');
    await page.keyboard.press('Escape');
    await expect(again).toHaveCount(0);
  });

  test('English, dark, 390px: the panel fits and an audit row hides the password', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/moderation/players');
    await expect(root(m).locator('[data-player-new]').first()).toHaveText('New player');
    await root(m).locator('[data-player-new]').first().click();
    await expect(m.locator('[data-player-form]')).toContainText('no confirmation email and no SMS are sent');
    await noHorizontalScroll(m);
    expect(Number(sql(`SELECT count(*) FROM console_audit_log WHERE metadata::text LIKE '%/publication/players%' AND create_time > now() - interval '10 minutes'`))).toBeGreaterThan(0);
    expect(sql(`SELECT count(*) FROM console_audit_log WHERE metadata::text LIKE '%e2e-pass-1234%'`)).toBe('0');
  });
});
