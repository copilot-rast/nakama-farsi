import crypto from 'node:crypto';
import { expect, test } from '../../lib/fixtures';
import { eco, uid } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { field, open, panel, pickRef, root, rows, go, savePanel, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const ids: Record<string, string> = {};

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const r of ['products', 'prizes', 'currencies']) {
    if (ids[r]) await eco.del('/' + r + '/' + ids[r]);
  }
  restoreConfig();
});

function serviceAccount() {
  const { privateKey } = crypto.generateKeyPairSync('rsa', { modulusLength: 2048 });
  return JSON.stringify({
    type: 'service_account', project_id: 'e2e', private_key_id: 'k1', client_id: '1',
    private_key: privateKey.export({ type: 'pkcs8', format: 'pem' }).toString(),
    client_email: 'play@e2e.iam.gserviceaccount.com', token_uri: 'https://oauth2.googleapis.com/token',
  });
}

function appleKey() {
  const { privateKey } = crypto.generateKeyPairSync('ec', { namedCurve: 'P-256' });
  return privateKey.export({ type: 'pkcs8', format: 'pem' }).toString();
}

async function method(name: string) {
  return (await eco.get('/payment-methods')).find((m: any) => m.method === name);
}

test.describe('store gateways: Google Play, Apple, Stripe', () => {
  test('payment methods page configures the three methods and masks their secrets', async ({ page }) => {
    const active: Record<string, boolean> = {};
    for (const m of await eco.get('/payment-methods')) active[m.method] = m.isActive;
    try {
      await open(page, '#/project-settings/payment-methods');
      const gp = root(page).locator('[data-method="google_play"]');
      const ap = root(page).locator('[data-method="apple"]');
      const st = root(page).locator('[data-method="stripe"]');
      await expect(gp).toContainText('گوگل‌پلی');
      await expect(ap).toContainText('اپ‌استور اپل');
      await expect(st).toContainText('استرایپ');
      for (const card of [gp, ap, st]) await expect(card).toContainText('تنظیم نشده');

      await field(gp, /^نام بسته$/).locator('input').fill('com.e2e.' + k.slice(4));
      const json = field(gp, /^کلید JSON حساب سرویس گوگل$/).locator('textarea');
      await expect(json).toHaveClass(/eco-secret/);
      await json.fill(serviceAccount());
      await gp.getByRole('button', { name: 'ذخیره' }).click();
      await expect(toasts(page).filter({ hasText: 'ذخیره شد.' }).last()).toBeVisible();
      const gpCard = root(page).locator('[data-method="google_play"]');
      await expect(gpCard).toContainText('تنظیم شده');
      await expect(field(gpCard, /^کلید JSON حساب سرویس گوگل$/).locator('textarea')).toHaveValue('********');

      await field(ap, /^Bundle ID$/).locator('input').fill('com.e2e.ios');
      await field(ap, /^Key ID$/).locator('input').fill('KEY' + k.slice(-4));
      await ap.getByRole('button', { name: 'ذخیره' }).click();
      await expect(ap.locator('.eco-notice, .eco-invalid').first()).toBeVisible();
      expect((await method('apple')).configured).toBe(false);
      await field(ap, /^Issuer ID$/).locator('input').fill('issuer-' + k.slice(-6));
      await field(ap, /^کلید خصوصی \.p8$/).locator('textarea').fill(appleKey());
      await field(ap, /^Shared secret/).locator('input').fill('apple-shared-' + k);
      await ap.getByRole('button', { name: 'ذخیره' }).click();
      await expect(toasts(page).filter({ hasText: 'ذخیره شد.' }).last()).toBeVisible();
      const apCard = root(page).locator('[data-method="apple"]');
      await expect(apCard).toContainText('تنظیم شده');
      await expect(field(apCard, /^کلید خصوصی \.p8$/).locator('textarea')).toHaveValue('********');
      await expect(field(apCard, /^Shared secret/).locator('input')).toHaveValue('********');

      await expect(field(st, /^واحد پول$/).locator('input')).toHaveValue('usd');
      await field(st, /^کلید مخفی/).locator('input').fill('sk_test_' + k);
      await field(st, /^رمز امضای وب‌هوک$/).locator('input').fill('whsec_' + k);
      await field(st, /^کلید عمومی \(Publishable key\)$/).locator('input').fill('pk_test_' + k);
      await st.getByRole('button', { name: 'ذخیره' }).click();
      await expect(field(st, /^کلید مخفی/)).toHaveClass(/eco-invalid/);
      expect((await method('stripe')).configured).toBe(false);
      await st.locator('[data-field="testMode"]').click();
      await st.getByRole('button', { name: 'ذخیره' }).click();
      await expect(toasts(page).filter({ hasText: 'ذخیره شد.' }).last()).toBeVisible();
      const stCard = root(page).locator('[data-method="stripe"]');
      await expect(stCard).toContainText('تنظیم شده');
      await expect(field(stCard, /^کلید مخفی/).locator('input')).toHaveValue('********');
      await expect(field(stCard, /^رمز امضای وب‌هوک$/).locator('input')).toHaveValue('********');

      const g = await method('google_play');
      const a = await method('apple');
      const s = await method('stripe');
      expect(g.config).toMatchObject({ packageName: 'com.e2e.' + k.slice(4), serviceAccountJson: '********', sandbox: false });
      expect(a.config).toMatchObject({ bundleId: 'com.e2e.ios', privateKeyP8: '********', sharedSecret: '********', issuerId: 'issuer-' + k.slice(-6) });
      expect(s.config).toMatchObject({ secretKey: '********', webhookSecret: '********', publishableKey: 'pk_test_' + k, currency: 'usd', testMode: true });
      expect(s.sandboxField).toBe('testMode');
      for (const m of [g, a, s]) expect(m.isActive).toBe(active[m.method]);

      await page.reload();
      await settle(page);
      const html = await page.content();
      expect(html).not.toContain('BEGIN PRIVATE KEY');
      expect(html).not.toContain('sk_test_' + k);
      expect(html).not.toContain('whsec_' + k);

      const card = root(page).locator('[data-method="stripe"]');
      await field(card, /^واحد پول$/).locator('input').fill('EUR');
      await card.getByRole('button', { name: 'ذخیره' }).click();
      await expect(toasts(page).filter({ hasText: 'ذخیره شد.' }).last()).toBeVisible();
      const after = await method('stripe');
      expect(after.config.currency).toBe('eur');
      expect(after.config.secretKey).toBe('********');
      expect(after.configured).toBe(true);
    } finally {
      restoreConfig();
    }
    const back = await eco.get('/payment-methods');
    for (const m of back.filter((x: any) => ['google_play', 'apple', 'stripe'].includes(x.method))) {
      expect(m.configured).toBe(false);
      expect(m.isActive).toBe(active[m.method]);
    }
  });

  test('product editor: store product ids and the Stripe price', async ({ page }) => {
    const cur = await eco.post('/currencies', { key: k.replace(/-/g, '_').slice(0, 30) + 'c', name: { fa: 'سکه ' + k }, type: 'soft' });
    ids.currencies = cur.id;
    const prize = await eco.post('/prizes', { key: k + 'p', name: { fa: 'جایزه ' + k }, isRepeatable: true, items: [{ type: 'currency', key: cur.key, action: 'add', value: 10 }] });
    ids.prizes = prize.id;
    await open(page, '#/economy/products/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill(k + 'g');
    await p.locator('[data-translated="name"] [role="tabpanel"][data-lang="fa"]').locator('input,textarea').first().fill('بستهٔ ' + k);
    await pickRef(p, /^جایزه$/, k + 'p');
    const box = (name: string) => p.locator('.x-checkbox', { hasText: name });
    await box('زرین‌پال').click();
    await expect(field(p, /^قیمت استرایپ/)).toBeHidden();
    await box('گوگل‌پلی').click();
    await box('اپ‌استور اپل').click();
    await box('استرایپ').click();
    await expect(field(p, /^شناسهٔ محصول گوگل‌پلی$/)).toBeVisible();
    await expect(field(p, /^شناسهٔ محصول اپل$/)).toBeVisible();
    await expect(field(p, /^قیمت استرایپ/)).toBeVisible();
    await field(p, /^شناسهٔ محصول گوگل‌پلی$/).locator('input').fill(k + '_gp');
    await field(p, /^شناسهٔ محصول اپل$/).locator('input').fill(k + '_ios');
    await field(p, /^قیمت استرایپ/).locator('input').fill('199');
    await field(p, /^واحد پول قیمت$/).locator('input').fill('USD');
    await savePanel(page);
    const list = await eco.get('/products?q=' + encodeURIComponent(k + 'g') + '&limit=5');
    const prod = list.items.find((x: any) => x.key === k + 'g');
    ids.products = prod.id;
    expect([...prod.paymentMethods].sort()).toEqual(['apple', 'google_play', 'stripe']);
    expect(prod).toMatchObject({ sku: { google_play: k + '_gp', apple: k + '_ios' }, priceMinor: 199, priceCurrency: 'usd' });
    await go(page, '#/economy/products?q=' + k + 'g');
    const row = rows(page).first();
    await expect(row).toContainText('گوگل‌پلی');
    await expect(row).toContainText('استرایپ');
  });

  test('payment settings page has the store and Stripe policies', async ({ page }) => {
    const before = (await eco.get('/payment-settings')).settings;
    expect(before).toMatchObject({ liveAcceptsTestPurchases: false, appStorePayload: 'optional', storeConsumed: 'accept', googleAcknowledge: true, stripeToleranceSeconds: 300 });
    try {
      await open(page, '#/economy/payment-settings');
      const f = root(page).locator('[data-settings-form="payments"]');
      await expect(f).toContainText('گوگل‌پلی، اپل و استرایپ');
      await field(f, /^بیشترین سن امضای وب‌هوک استرایپ/).locator('input').fill('600');
      await f.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toBeVisible();
      expect((await eco.get('/payment-settings')).settings.stripeToleranceSeconds).toBe(600);
    } finally {
      await eco.put('/payment-settings', before);
    }
  });
});
