import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';
import { field, openSettings, pickRef, root, settle, toasts } from '../../lib/console';

const k = uid();
const cur = ('un' + k.replace(/[^a-z0-9]/g, '')).slice(0, 24);
let curId = '';
let P: Player;
let savedName = '';
let savedCode = '';

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

function refusal(res: { status: number; data: any }) {
  const msg = typeof res.data === 'string' ? res.data : res.data?.message || res.data?.error || '';
  try { return JSON.parse(msg); } catch { return { code: msg }; }
}

test.beforeAll(async () => {
  savedName = sql(`SELECT COALESCE((SELECT value::text FROM base_username_settings WHERE id=1), '')`);
  savedCode = sql(`SELECT COALESCE((SELECT value::text FROM base_player_code_settings WHERE id=1), '')`);
  const c = await eco.post('/currencies', { key: cur, name: { fa: 'جم نام ' + k, en: 'Name gem ' + k }, type: 'hard', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  curId = c.id;
  P = await devicePlayer('e2e-un-' + k, ('un' + k).replace(/_/g, '').slice(0, 20));
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  if (savedName) sql(`UPDATE base_username_settings SET value=${q(savedName)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_username_settings`);
  if (savedCode) sql(`UPDATE base_player_code_settings SET value=${q(savedCode)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_player_code_settings`);
  if (P) await deleteUser(P.id);
  if (curId) await eco.del('/currencies/' + curId);
});

test.describe('username change policy', () => {
  test('the accounts tab saves the price ladder, cooldown and player code settings', async ({ page }) => {
    await openSettings(page, 'accounts');
    const form = root(page).locator('[data-settings-form="username-policy"]');
    await expect(form).toContainText('تغییر نام کاربری');
    await expect(form).toContainText('پله‌های قیمت');
    await form.locator('label', { hasText: 'قانون تغییر نام روشن باشد' }).click();
    await pickRef(form, 'ارز هزینه', cur);
    await field(form, 'پله‌های قیمت').locator('input').fill('۵، 10');
    await field(form, 'فاصله بین دو تغییر (ساعت)').locator('input').fill('0');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    await expect.poll(() => sql(`SELECT value->>'enabled' || '|' || (value->'ladder')::text || '|' || (value->>'currency') FROM base_username_settings`)).toBe('true|[5, 10]|' + cur);
    await field(form, 'پله‌های قیمت').locator('input').fill('0, 5');
    await form.locator('[data-settings-save]').click();
    await expect(form.locator('.eco-invalid')).toContainText('حداکثر بیست عدد');

    const codes = root(page).locator('[data-settings-form="player-codes"]');
    await expect(codes).toContainText('کد بازیکن');
    await field(codes, 'طول کد').locator('input').fill('7');
    await codes.locator('[data-settings-save]').click();
    await expect.poll(() => sql(`SELECT value->>'length' FROM base_player_code_settings`)).toBe('7');
    await base.put('/player-codes/settings', { length: 6, charset: 'alnum' });
  });

  test('first rename is free, then the ladder price, price check, not enough gems, and UpdateAccount pays too', async () => {
    const quote = await rpc(P, 'base_username_quote', {});
    expect(quote.data).toMatchObject({ enabled: true, free: true, cost: 0, currency: cur });
    const first = await rpc(P, 'base_username_change', { username: 'one' + k.slice(-6) });
    expect(first.status).toBe(200);
    expect(first.data).toMatchObject({ source: 'free', cost: 0, quote: { free: false, cost: 5 } });

    const poor = await rpc(P, 'base_username_change', { username: 'two' + k.slice(-6) });
    expect(poor.status).toBe(400);
    expect(refusal(poor)).toMatchObject({ code: 'CURRENCY_IS_NOT_ENOUGH', lang: 'fa', meta: { amount: 5, currencyKey: cur } });

    sql(`UPDATE users SET wallet = jsonb_set(COALESCE(wallet,'{}'::jsonb), '{${cur}}', '30') WHERE id='${P.id}'`);
    const stale = await rpc(P, 'base_username_change', { username: 'two' + k.slice(-6), expectedCost: 99 });
    expect(refusal(stale)).toMatchObject({ code: 'USERNAME_PRICE_CHANGED', meta: { cost: 5 } });
    const paid = await rpc(P, 'base_username_change', { username: 'two' + k.slice(-6), expectedCost: 5 });
    expect(paid.data).toMatchObject({ source: 'paid', cost: 5, currency: cur, quote: { cost: 10, paidChanges: 1 } });

    await updateAccount(P, { username: 'three' + k.slice(-6) });
    expect(sql(`SELECT username || '|' || (wallet->>'${cur}') FROM users WHERE id='${P.id}'`)).toBe('three' + k.slice(-6) + '|15');
    const same = await rpc(P, 'base_username_change', { username: 'three' + k.slice(-6) });
    expect(refusal(same).code).toBe('USERNAME_UNCHANGED');

    await base.put('/username-settings', { enabled: true, firstFree: true, currency: cur, ladder: [5, 10], cooldownHours: 3 });
    const wait = await rpc(P, 'base_username_change', { username: 'four' + k.slice(-6) });
    const r = refusal(wait);
    expect(r.code).toBe('USERNAME_CHANGE_COOLDOWN');
    expect(r.meta.cooldownEndsAt).toBeTruthy();
    expect(r.message).toMatch(/[۰-۹]/);
  });

  test('the player card shows the code and history and staff rename skips price and cooldown', async ({ page }) => {
    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-social]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    const code = sql(`SELECT code FROM base_player_code WHERE user_id='${P.id}'`);
    await expect(card.locator('[data-ps-code]')).toHaveText(code);
    await expect(card.locator('[data-un-history]')).toHaveAttribute('data-un-history', '3');
    await expect(card.locator('[data-un-quote]')).toHaveAttribute('data-un-quote', 'paid');
    await expect(card).toContainText('تا ');
    await card.locator('[data-ps-action="rename"]').click();
    const dlg = page.locator('[data-form-dialog="player-rename"]');
    await dlg.locator('[data-rename-input]').fill('staff' + k.slice(-6));
    await dlg.locator('[data-dialog-ok]').click();
    await expect(dlg).toHaveCount(0);
    await expect(card.locator('[data-un-history]')).toHaveAttribute('data-un-history', '4', { timeout: 20_000 });
    await expect(card).toContainText('کنسول');
    expect(sql(`SELECT username || '|' || (wallet->>'${cur}') FROM users WHERE id='${P.id}'`)).toBe('staff' + k.slice(-6) + '|15');
    expect(sql(`SELECT source || '|' || admin FROM base_username_change WHERE user_id='${P.id}' ORDER BY created_at DESC LIMIT 1`)).toMatch(/^admin\|.+/);
    expect(Number(sql(`SELECT count(*) FROM base_activity WHERE user_id='${P.id}' AND kind='account'`))).toBe(4);

    await card.locator('[data-ps-action="rename"]').click();
    await dlg.locator('[data-rename-input]').fill('staff' + k.slice(-6));
    await dlg.locator('[data-dialog-ok]').click();
    await expect(dlg).toContainText('نام کاربری تازه با نام فعلی یکی است');
    await settle(page, card);
  });
});
