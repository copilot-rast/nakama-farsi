import fs from 'node:fs';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, must, uid } from '../../lib/api';
import { loginAs, noHorizontalScroll, open, root, settle, toasts } from '../../lib/console';
import { restoreConfig, sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';

const PATH = '/config-transfer';
const k = uid().replace(/-/g, '_');
const keep = 'e2e_ct_keep_' + k.slice(-6);
const skip = 'e2e_ct_skip_' + k.slice(-6);
let exported: any = null;
let coinSort = '';

test.describe.configure({ mode: 'serial' });

test.beforeAll(() => {
  coinSort = sql(`SELECT sort FROM economy_currency WHERE key='coin'`);
});

test.afterAll(() => {
  sql(`DELETE FROM economy_currency WHERE key IN ('${keep}','${skip}')`);
  if (coinSort !== '') sql(`UPDATE economy_currency SET sort=${Number(coinSort)} WHERE key='coin'`);
  restoreConfig();
});

function currencyRow(key: string, sort: number) {
  return { key, values: { key, name: { fa: 'ارز آزمایشی', en: 'Test coin' }, type: 'soft', default_value: 3, icon_url: '', metadata: {}, sort, category_id: null, tag_ids: [], regen: null } };
}

test('the API lists the areas, exports by keys without secrets and is admin only', async () => {
  const d = await base.get(PATH);
  expect(d.admin).toBe(true);
  expect(d.format).toBe('nakama-config');
  expect(d.settings.maxFileMB).toBe(20);
  expect(d.defaults.maxFileMB).toBe(20);
  const keys = d.areas.map((a: any) => a.key);
  for (const a of ['taxonomy', 'currencies', 'metas', 'prizes', 'items', 'products', 'tasks', 'achievements', 'streaks', 'referral', 'games', 'leaderboards', 'payment_methods', 'app_versions', 'server_rules']) expect(keys).toContain(a);
  const f = await must(consoleApi('POST', '/v2/console/base' + PATH + '/export', { areas: ['currencies', 'payment_methods', 'products'] }));
  expect(f.format).toBe('nakama-config');
  expect(f.version).toBe(1);
  expect(f.source.exportedAt).toBeTruthy();
  expect(f.areas.map((a: any) => a.key).sort()).toEqual(['currencies', 'payment_methods', 'products']);
  const text = JSON.stringify(f);
  for (const name of ['merchantId', 'merchant', 'apiSecret', 'accessToken', 'secretKey', 'webhookSecret', 'serviceAccountJson', 'privateKeyP8', 'sharedSecret']) expect(text).not.toContain('"' + name + '"');
  const product = f.areas.find((a: any) => a.key === 'products').tables[0].rows[0];
  if (product) {
    expect(product.values.id).toBeUndefined();
    expect(product.values.prize_id).not.toMatch(/^[0-9a-f]{8}-/);
  }
  const bad = await consoleApi('POST', '/v2/console/base' + PATH + '/export', { areas: ['nope'] });
  expect(bad.status).toBe(400);
  expect(bad.data.error).toBe('CONFIG_AREA_UNKNOWN');
  const notFile = await consoleApi('POST', '/v2/console/base' + PATH + '/preview', { file: { format: 'other' } });
  expect(notFile.data.error).toBe('CONFIG_FILE_INVALID');
  const same = await must(consoleApi('POST', '/v2/console/base' + PATH + '/preview', { file: f }));
  expect(same.result.applied).toBe(false);
  expect(same.result.totals.new + same.result.totals.changed).toBe(0);
});

test('the page is in Ctrl K and downloads the export file', async ({ page }) => {
  await open(page, '#/system/updates');
  await page.keyboard.press('Control+k');
  await page.locator('.pl-pal-input').fill('انتقال تنظیمات');
  await expect(page.locator('.pl-pal-row').first()).toContainText('انتقال تنظیمات');
  await page.keyboard.press('Enter');
  await expect(page).toHaveURL(/#\/system\/config-transfer/);
  await settle(page);
  const box = root(page).locator('[data-ct-page]');
  await expect(box.locator('[data-ct-export-card] h2')).toHaveText('خروجی گرفتن');
  await expect(box.locator('[data-ct-import-card] h2')).toHaveText('وارد کردن');
  await expect(box.locator('[data-ct-export-count]')).toContainText('بخش');
  await box.locator('[data-ct-export-all]').click();
  await expect(box.locator('[data-ct-export]')).toBeDisabled();
  await box.locator('[data-ct-area="currencies"]').click();
  await box.locator('[data-ct-area="products"]').click();
  await box.locator('[data-ct-area="prizes"]').click();
  await expect(box.locator('[data-ct-export-count]')).toContainText('۳');
  const wait = page.waitForEvent('download');
  await box.locator('[data-ct-export]').click();
  const dl = await wait;
  expect(dl.suggestedFilename()).toMatch(/^config-.*\.json$/);
  exported = JSON.parse(fs.readFileSync(await dl.path(), 'utf8'));
  expect(exported.areas.map((a: any) => a.key).sort()).toEqual(['currencies', 'prizes', 'products']);
  expect(exported.counts.currencies).toBeGreaterThan(0);
  await expect(toasts(page).last()).toContainText('دانلود شد');
  await expect(box.locator('[data-ct-runs] [data-ct-run="export"]').first()).toBeVisible();
});

test('an admin previews an import, leaves one row out and applies it', async ({ page }) => {
  expect(exported).toBeTruthy();
  const file = JSON.parse(JSON.stringify(exported));
  const cur = file.areas.find((a: any) => a.key === 'currencies').tables[0];
  const coin = cur.rows.find((r: any) => r.key === 'coin');
  coin.values.sort = Number(coinSort) + 77;
  cur.rows.push(currencyRow(keep, 5), currencyRow(skip, 6));
  await open(page, '#/system/config-transfer');
  const box = root(page).locator('[data-ct-page]');
  await box.locator('[data-ct-input]').setInputFiles({ name: 'dev.json', mimeType: 'application/json', buffer: Buffer.from(JSON.stringify(file)) });
  const pv = box.locator('[data-ct-preview-result="preview"]');
  await expect(pv).toBeVisible();
  await expect(box.locator('[data-ct-file-name]')).toContainText('dev.json');
  await expect(pv.locator('[data-ct-source]')).toContainText('از سرور');
  await expect(pv.locator('[data-ct-stat="new"]')).toContainText('۲');
  await expect(pv.locator('[data-ct-stat="changed"]')).toContainText('۱');
  await expect(pv.locator('[data-ct-area-row="currencies"]')).toBeVisible();
  const coinRow = pv.locator('[data-ct-row="economy_currency:coin"]');
  await expect(coinRow).toHaveAttribute('data-status', 'changed');
  await expect(pv.locator('tr').filter({ has: page.locator('[data-ct-row="economy_currency:coin"]') }).locator('[data-ct-change="sort"]')).toBeVisible();
  await pv.locator('[data-ct-filter="status"] select').selectOption('new');
  await expect(pv.locator('[data-ct-row]')).toHaveCount(2);
  await pv.locator(`[data-economy-sel="ct-rows"][data-id="currencies|economy_currency:${skip}"]`).click();
  await expect(pv.locator('[data-ct-selbar]')).toBeVisible();
  await pv.locator('[data-ct-selbar]').getByRole('button', { name: 'اعمال نشوند' }).click();
  await expect(pv.locator(`[data-ct-row="economy_currency:${skip}"]`)).toHaveAttribute('data-excluded', 'true');
  expect(sql(`SELECT count(*) FROM economy_currency WHERE key='${keep}'`)).toBe('0');

  await pv.locator('[data-ct-apply]').click();
  const dialog = page.locator('[role="alertdialog"]');
  await expect(dialog).toContainText('تغییرات اعمال شود؟');
  await expect(dialog).toContainText('یک تراکنش');
  await dialog.getByRole('button', { name: 'اعمال کن' }).click();
  await expect(toasts(page).last()).toContainText('تنظیمات وارد شد');
  const done = box.locator('[data-ct-preview-result="applied"]');
  await expect(done).toBeVisible();
  await expect(done.locator(`[data-ct-row="economy_currency:${skip}"]`)).toHaveAttribute('data-status', 'skipped');
  expect(sql(`SELECT count(*) FROM economy_currency WHERE key='${keep}'`)).toBe('1');
  expect(sql(`SELECT count(*) FROM economy_currency WHERE key='${skip}'`)).toBe('0');
  expect(sql(`SELECT sort FROM economy_currency WHERE key='coin'`)).toBe(String(Number(coinSort) + 77));
  await expect(box.locator('[data-ct-runs] [data-ct-run="import"]').first()).toBeVisible();
  expect(sql(`SELECT count(*) FROM console_audit_log WHERE message='base config transfer apply' AND create_time > now() - interval '5 minutes'`)).not.toBe('0');

  const again = await must(consoleApi('POST', '/v2/console/base' + PATH + '/preview', { file }));
  const rows = again.result.areas.flatMap((a: any) => a.rows);
  expect(rows.find((r: any) => r.key === keep).status).toBe('same');
  expect(rows.find((r: any) => r.key === 'coin').status).toBe('same');
  expect(rows.find((r: any) => r.key === skip).status).toBe('new');
});

test('the settings panel saves the file size limit', async ({ page }) => {
  await open(page, '#/system/config-transfer');
  await root(page).locator('[data-ct-open-settings]').click();
  const panel = page.locator('[data-ct-settings]');
  await expect(panel).toBeVisible();
  await panel.locator('[data-ct-set="maxFileMB"]').fill('0');
  await page.locator('[data-ct-save]').click();
  await expect(page.locator('.eco-invalid')).toBeVisible();
  await panel.locator('[data-ct-set="maxFileMB"]').fill('25');
  await panel.locator('[data-ct-set="serverName"]').fill('سرور آزمایشی');
  await page.locator('[data-ct-save]').click();
  await expect(toasts(page).last()).toContainText('ذخیره شد');
  const d = await base.get(PATH);
  expect(d.settings.maxFileMB).toBe(25);
  expect(d.serverName).toBe('سرور آزمایشی');
  await base.put(PATH + '/settings', d.defaults);
});

test('a settings reader sees the page but cannot export or import', async ({ spawn }) => {
  const who = { user: 'e2ect' + k.slice(-8), pass: 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!' };
  const add = await consoleApi('POST', '/v2/console/user', { username: who.user, email: who.user + '@example.com', acl: { ACCOUNT: { read: true }, SETTINGS: { read: true, write: true } } });
  expect(add.status).toBeLessThan(300);
  try {
    const reset = await must(consoleApi('POST', '/v2/console/user/' + who.user + '/reset/password'));
    const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: reset.code || add.data?.token, password: who.pass }) });
    expect(pc.ok).toBe(true);
    const auth = await (await fetch(CONSOLE_URL + '/v2/console/authenticate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username: who.user, password: who.pass }) })).json();
    const bearer = 'Bearer ' + auth.token;
    const view = await consoleApi('GET', '/v2/console/base' + PATH, undefined, bearer);
    expect(view.status).toBe(200);
    expect(view.data.admin).toBe(false);
    for (const p of ['/export', '/preview', '/apply']) {
      const r = await consoleApi('POST', '/v2/console/base' + PATH + p, {}, bearer);
      expect(r.status, p).toBe(403);
    }
    const p = await spawn({ signedIn: false });
    await loginAs(p, who);
    await p.goto(CONSOLE_URL + '/#/system/config-transfer');
    await settle(p);
    await expect(root(p).locator('[data-ct-admin-only]')).toContainText('فقط مدیر کنسول');
    await expect(root(p).locator('[data-ct-export]')).toBeDisabled();
    await expect(root(p).locator('[data-ct-pick]')).toBeDisabled();
  } finally {
    await consoleApi('DELETE', '/v2/console/user?username=' + who.user);
  }
});

test('English and dark mobile layout', async ({ spawn }) => {
  const p = await spawn({ signedIn: true, locale: 'en', mode: 'dark', mobile: true, width: 390, height: 844 });
  await open(p, '#/system/config-transfer');
  const box = root(p).locator('[data-ct-page]');
  await expect(box.locator('[data-ct-export-card] h2')).toHaveText('Export');
  await expect(box.locator('[data-ct-import-card] h2')).toHaveText('Import');
  await box.locator('[data-ct-input]').setInputFiles({ name: 'x.json', mimeType: 'application/json', buffer: Buffer.from('{"format":"x"}') });
  await expect(box.locator('[data-ct-error="CONFIG_FILE_INVALID"]')).toContainText('not a settings transfer file');
  await noHorizontalScroll(p);
});
