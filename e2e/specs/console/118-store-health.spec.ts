import http from 'node:http';
import { expect, test } from '../../lib/fixtures';
import { consoleApi, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { open, openSettings, psWindow, toasts } from '../../lib/console';
import { sql } from '../../lib/db';

const FAKE_PORT = Number(process.env.E2E_FAKE_BAZAAR_PORT || 7598);
const k = uid('sh').replace(/[^a-z0-9_]/g, '');
const sku = 'fir_oil_' + k;
const reply = { status: 403, body: '{"error":"access_forbidden","error_description":"Invalid API secret"}' };
const seen: string[] = [];
let fake: http.Server;
let saved = '';
let productId = '';
let prizeId = '';
let player: Player;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async () => {
  fake = http.createServer((req, res) => {
    seen.push(req.url || '');
    res.writeHead(reply.status, { 'Content-Type': 'application/json' });
    res.end(reply.body);
  });
  await new Promise<void>((resolve) => fake.listen(FAKE_PORT, '127.0.0.1', () => resolve()));
  saved = sql(`SELECT COALESCE((SELECT is_active::text || '|' || config FROM economy_payment_method WHERE method='bazaar'), '')`);
  await eco.put('/payment-methods/bazaar', { isActive: true, config: { sandbox: false, packageName: 'ir.e2e.roadstorm', apiSecret: 'E2E-WRONG-SECRET' } });
  prizeId = (await eco.post('/prizes', { key: 'p_' + k, name: { fa: 'سوخت', en: 'Fuel' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 100 }] })).id;
  productId = (await eco.post('/products', { key: 'pr_' + k, isActive: true, name: { fa: 'سوخت', en: 'Fuel' }, prizeId, paymentMethods: ['bazaar'], priceToman: 20000, sku: { bazaar: sku } })).id;
  player = await devicePlayer('e2e-sh-' + k);
});

test.afterAll(async () => {
  fake?.close();
  if (player) {
    sql(`DELETE FROM economy_payment_log WHERE payment_id IN (SELECT id FROM economy_payment WHERE user_id='${player.id}')`);
    sql(`DELETE FROM economy_user_prize WHERE user_id='${player.id}'`);
    sql(`DELETE FROM economy_payment WHERE user_id='${player.id}'`);
    await deleteUser(player.id);
  }
  if (productId) await eco.del('/products/' + productId);
  if (prizeId) await eco.del('/prizes/' + prizeId);
  if (saved) {
    const [active, config] = [saved.slice(0, saved.indexOf('|')), saved.slice(saved.indexOf('|') + 1)];
    sql(`UPDATE economy_payment_method SET is_active=${active === 'true'}, config='${config.replace(/'/g, "''")}', store_last_failure=NULL, store_last_ok_at=NULL, store_last_test=NULL WHERE method='bazaar'`);
  } else {
    sql(`DELETE FROM economy_payment_method WHERE method='bazaar'`);
  }
});

let paymentId = '';

test('a rejected Bazaar secret is a permanent error and the payment stays pending', async () => {
  const start = await rpc(player, 'economy_purchase_start', { productId, method: 'bazaar' });
  expect(start.status).toBe(200);
  paymentId = start.data.paymentId;
  const submit = await rpc(player, 'economy_purchase_submit', { paymentId, token: 'real-token-' + k });
  expect(submit.status).not.toBe(503);
  expect(JSON.stringify(submit.data)).toContain('STORE_CONFIGURATION_REJECTED');
  expect(submit.data.code).toBe(9);
  expect(sql(`SELECT status || '|' || error FROM economy_payment WHERE id='${paymentId}'`)).toBe('pending|STORE_CONFIGURATION_REJECTED');
  const review = await rpc(player, 'economy_purchase_review', { method: 'bazaar', sku, token: 'probe-not-a-real-token-000' });
  expect(review.data.code).toBe(9);
  const list = await eco.get('/payment-methods');
  const bz = list.find((m: any) => m.method === 'bazaar');
  expect(bz.health.lastFailure.httpStatus).toBe(403);
  expect(bz.health.lastFailure.storeError).toBe('access_forbidden');
  expect(JSON.stringify(list)).not.toContain('E2E-WRONG-SECRET');
});

test('the payment method card shows the last failure and tests the credentials', async ({ page }) => {
  await openSettings(page, 'payment-methods');
  const card = psWindow(page).locator('[data-method="bazaar"]');
  const health = card.locator('[data-store-health="bazaar"]');
  await expect(health).toBeVisible();
  await expect(health).toContainText('بررسی کلیدهای فروشگاه');
  await expect(health.locator('[data-last-failure]')).toContainText('access_forbidden');
  await expect(health.locator('[data-last-failure]')).toContainText('۴۰۳');
  await expect(health.locator('[data-last-failure]')).toContainText('کلید رد شد');

  await health.locator('[data-store-test="bazaar"]').click();
  await expect(health.locator('[data-result="credentials_rejected"]')).toBeVisible();
  await health.locator('[data-store-test-sku]').fill(sku);
  await health.locator('[data-store-test="bazaar"]').click();
  await expect(health.locator('[data-store-test-result]')).toContainText(sku);

  reply.status = 404;
  reply.body = '{"error":"invalid_value","error_description":"Package name is invalid"}';
  await health.locator('[data-store-test="bazaar"]').click();
  await expect(health.locator('[data-result="package_not_found"]')).toBeVisible();

  reply.body = '{"error":"not_found","error_description":"The requested purchase is not found!"}';
  await health.locator('[data-store-test="bazaar"]').click();
  await expect(health.locator('[data-result="credentials_ok"]')).toBeVisible();
  await expect(toasts(page).filter({ hasText: 'کلیدها درست است' })).toBeVisible();
  expect(seen[seen.length - 1]).toContain('/api/validate/ir.e2e.roadstorm/inapp/' + sku + '/purchases/nakama-credentials-test-');
  await expect(health.locator('[data-last-ok]')).toBeVisible();
  expect(await page.content()).not.toContain('E2E-WRONG-SECRET');
});

test('the payment detail shows the store answer and re-checks after the fix', async ({ page }) => {
  await open(page, '#/economy/payments/' + paymentId);
  const answer = page.locator('.eco-dl').filter({ hasText: 'آخرین پاسخ فروشگاه' });
  await expect(answer).toContainText('access_forbidden');
  reply.status = 200;
  reply.body = JSON.stringify({ purchaseState: 0, consumptionState: 1, developerPayload: paymentId, purchaseTime: 1 });
  await page.locator('[data-payment-recheck]').click();
  await expect(toasts(page).filter({ hasText: 'فروشگاه خرید را تأیید کرد.' })).toBeVisible();
  await expect(page.locator('[data-payment-recheck]')).toHaveCount(0);
  expect(sql(`SELECT status || '|' || reward_granted::text FROM economy_payment WHERE id='${paymentId}'`)).toBe('success|true');
  const again = await rpc(player, 'economy_purchase_review', { method: 'bazaar', sku, token: 'real-token-' + k });
  expect(again.status).toBe(200);
  expect(again.data.status).toBe('success');
  expect(again.data.paymentId).toBe(paymentId);
});

test('Ctrl+K finds the store credentials test', async ({ page }) => {
  await open(page, '#/economy/payments');
  await page.keyboard.press('Control+k');
  const input = page.locator('.pl-pal-input');
  await input.fill('آزمودن کلیدهای');
  const row = page.locator('.pl-pal-row').filter({ hasText: 'آزمودن کلیدهای فروشگاه' });
  await expect(row).toHaveCount(1);
  await page.keyboard.press('Enter');
  await expect(psWindow(page).locator('[data-store-health="bazaar"]')).toBeVisible();
  expect((await consoleApi('POST', '/v2/console/economy/payment-methods/zarinpal/test', {})).status).toBe(400);
});
