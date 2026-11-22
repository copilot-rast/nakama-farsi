import { expect, test } from '../../lib/fixtures';
import { consoleApi, deleteUser, devicePlayer, must, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { API_URL, CONSOLE_URL } from '../../lib/env';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';

const k = uid();
const DK = '/v2/console/base/devkit';
const dk = {
  get: <T = any>(p: string) => must<T>(consoleApi('GET', DK + p)),
  put: <T = any>(p: string, b: unknown) => must<T>(consoleApi('PUT', DK + p, b)),
  post: <T = any>(p: string, b: unknown) => must<T>(consoleApi('POST', DK + p, b)),
  del: (p: string) => consoleApi('DELETE', DK + p),
};
const CUSTOM = ('E2E_' + k.replace(/^e2e_/, '') + '_LOCKED').toUpperCase();
const SCOPE = 'e2e' + k.slice(-4).replace(/[^a-z0-9]/g, 'x') + ':write';
let P: Player;
let original: any;

test.describe.configure({ mode: 'serial' });

function devkitRoot(page: import('@playwright/test').Page) {
  return page.locator('#devkit-root');
}

async function openDevkit(page: import('@playwright/test').Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/devkit/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(devkitRoot(page).locator(`[data-devkit-page="${sub}"]`)).toBeVisible();
}

function errEnvelope(res: { data: any }) {
  return JSON.parse(res.data.message);
}

async function keyCall(path: string, key?: string) {
  const r = await fetch(API_URL + path, { headers: key ? { 'X-Api-Key': key } : {} });
  return { status: r.status, data: await r.json().catch(() => null), headers: r.headers };
}

test.beforeAll(async () => {
  original = (await dk.get('/settings')).settings;
  P = await devicePlayer('e2e-dk-' + k, ('dk' + k).replace(/_/g, '').slice(0, 24));
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  const { updatedBy, updatedAt, ...clean } = original || {};
  if (original) await dk.put('/settings', clean);
  sql(`DELETE FROM devkit_error_message WHERE code IN ('GIFT_EXPIRED','${CUSTOM}'); DELETE FROM devkit_api_key WHERE name LIKE 'e2e %${k.slice(-6)}'; DELETE FROM devkit_api_scope WHERE name = '${SCOPE}';`);
  if (P) await deleteUser(P.id);
});

test.describe('devkit: error messages', () => {
  test('sidebar link, list with fa/en defaults, edit a text, catalog and envelope follow, reset to default', async ({ page }) => {
    await openDevkit(page, 'errors');
    await expect(page.locator('a[href="#/devkit/api-keys"]').first()).toBeAttached();
    const root = devkitRoot(page);
    await root.locator('input[name="q"]').fill('GIFT_EXPIRED');
    const row = root.locator('tr[data-devkit-row="GIFT_EXPIRED"]');
    await expect(row).toBeVisible();
    await expect(row.locator('[data-devkit-text="fa"]')).toHaveText('مهلت این کد هدیه تمام شده است.');
    await expect(row.locator('[data-devkit-text="en"]')).toHaveText('This gift code has expired.');
    await row.click();
    const panel = page.locator('[data-devkit-panel="error"]');
    await expect(panel).toBeVisible();
    await panel.locator('textarea[name="message-fa"]').fill('کد هدیه سوخت ' + k.slice(-4));
    await panel.locator('[data-devkit-action="save-error"]').click();
    await expect(panel).toHaveCount(0);
    await expect(row.locator('[data-devkit-text="fa"]')).toHaveText('کد هدیه سوخت ' + k.slice(-4));
    await expect(row).toContainText('ویرایش‌شده');

    await expect.poll(async () => {
      const cat = await rpc(P, 'devkit_error_catalog', { lang: 'fa', codes: ['GIFT_EXPIRED'] });
      return cat.data.errors.GIFT_EXPIRED.message;
    }, { timeout: 20_000 }).toBe('کد هدیه سوخت ' + k.slice(-4));
    const pubCat = await keyCall('/v2/devkit/errors?lang=en&codes=GIFT_EXPIRED');
    expect(pubCat.status).toBe(200);
    expect(pubCat.data.errors.GIFT_EXPIRED).toEqual({ status: 409, message: 'This gift code has expired.' });

    await row.click();
    await page.locator('[data-devkit-action="reset-error"]').click();
    await expect(page.locator('[data-devkit-panel="error"]')).toHaveCount(0);
    await expect(row.locator('[data-devkit-text="fa"]')).toHaveText('مهلت این کد هدیه تمام شده است.');
    expect(sql(`SELECT count(*) FROM devkit_error_message WHERE code='GIFT_EXPIRED'`)).toBe('0');
  });

  test('a custom code with a placeholder is created from the console and can be deleted', async ({ page }) => {
    await openDevkit(page, 'errors');
    const root = devkitRoot(page);
    await root.locator('[data-devkit-action="new-error"]').click();
    const panel = page.locator('[data-devkit-panel="error"]');
    await panel.locator('input[name="code"]').fill('bad code');
    await panel.locator('textarea[name="message-fa"]').fill('از سطح {unlockLevel} باز می‌شود');
    await panel.locator('[data-devkit-action="save-error"]').click();
    await expect(panel).toContainText('INVALID_CODE');
    await panel.locator('input[name="code"]').fill(CUSTOM);
    await panel.locator('textarea[name="message-en"]').fill('Unlocks at level {unlockLevel}');
    await panel.locator('input[name="status"]').fill('423');
    await panel.locator('[data-devkit-action="save-error"]').click();
    await expect(panel).toHaveCount(0);
    await root.locator('input[name="q"]').fill(CUSTOM);
    const row = root.locator(`tr[data-devkit-row="${CUSTOM}"]`);
    await expect(row).toContainText('کد سفارشی');
    await expect(row).toContainText('423');
    const entry = await dk.get('/errors/' + CUSTOM);
    expect(entry).toMatchObject({ custom: true, status: 423, messages: { fa: 'از سطح {unlockLevel} باز می‌شود', en: 'Unlocks at level {unlockLevel}' } });
    await row.click();
    await page.locator('[data-devkit-action="delete-error"]').click();
    await page.locator('[data-devkit-confirm="ok"]').click();
    await expect(row).toHaveCount(0);
    expect((await consoleApi('GET', DK + '/errors/' + CUSTOM)).status).toBe(404);
  });

  test('settings: code mode makes new RPCs return the bare code; envelope mode returns the localized envelope; module mode is saved', async ({ page }) => {
    await openDevkit(page, 'errors');
    const box = devkitRoot(page).locator('[data-devkit-settings="errors"]');
    await box.locator('[data-devkit-settings-toggle]').click();
    await box.locator('select[name="mode"]').selectOption('code');
    await box.locator('[data-devkit-settings-save]').click();
    await expect(page.locator('.eco-toasts .eco-notice').last()).toBeVisible();
    expect((await dk.get('/settings')).settings.errors.mode).toBe('code');
    await expect.poll(async () => (await rpc(P, 'devkit_rpc_schema', { module: 'nope' })).data?.message, { timeout: 20_000 }).toBe('DEVKIT_UNKNOWN_MODULE');
    await box.locator('select[name="mode"]').selectOption('envelope');
    await box.locator('[data-devkit-settings-save]').click();
    await expect.poll(async () => {
      const r = await rpc(P, 'devkit_rpc_schema', { module: 'nope' });
      return r.data?.message?.startsWith('{') ? errEnvelope(r) : null;
    }, { timeout: 20_000 }).toMatchObject({ code: 'DEVKIT_UNKNOWN_MODULE', status: 404, message: 'ماژولی به نام nope وجود ندارد.', meta: { module: 'nope' } });
    const moduleMode = box.locator('select[name="moduleMode"]');
    await expect(moduleMode).toHaveValue('code');
    await moduleMode.selectOption('envelope');
    await box.locator('[data-devkit-settings-save]').click();
    await expect.poll(async () => (await dk.get('/settings')).settings.errors.moduleMode).toBe('envelope');
    await moduleMode.selectOption('code');
    await box.locator('[data-devkit-settings-save]').click();
    await expect.poll(async () => (await dk.get('/settings')).settings.errors.moduleMode).toBe('code');
  });
});

test.describe('devkit: api keys', () => {
  let full = '';

  test('create a scoped key from the console, shown once, verified by the helper, revoke and delete', async ({ page }) => {
    await dk.post('/api-scopes', { name: SCOPE, description: 'e2e scope' });
    await openDevkit(page, 'api-keys');
    const root = devkitRoot(page);
    await expect(root.locator(`[data-devkit-list="api-scopes"] tr[data-devkit-row="${SCOPE}"]`)).toBeVisible();
    await root.locator('[data-devkit-action="new-key"]').click();
    const panel = page.locator('[data-devkit-panel="key"]');
    await panel.locator('[data-devkit-action="save-key"]').click();
    await expect(panel).toContainText('NAME_REQUIRED');
    await panel.locator('input[name="name"]').fill('e2e editor ' + k.slice(-6));
    await panel.locator('[data-devkit-action="save-key"]').click();
    await expect(panel).toContainText('SCOPES_REQUIRED');
    await panel.locator('input[data-devkit-scope="schema:read"]').check();
    await panel.locator('textarea[name="allowedIps"]').fill('not-an-ip');
    await panel.locator('[data-devkit-action="save-key"]').click();
    await expect(panel).toContainText('INVALID_IP');
    await panel.locator('textarea[name="allowedIps"]').fill('127.0.0.1\n::1');
    await panel.locator('select[name="expiryMode"]').selectOption('days');
    await panel.locator('input[name="expiresInDays"]').fill('7');
    await panel.locator('[data-devkit-action="save-key"]').click();
    const shown = page.locator('[data-devkit-panel="key-created"]');
    await expect(shown).toContainText('فقط همین یک بار');
    full = await shown.locator('[data-devkit-created-key]').inputValue();
    expect(full).toMatch(/^nfk_[a-z0-9]{12}_[A-Za-z0-9_-]{43}$/);
    await shown.locator('[data-devkit-action="key-done"]').click();
    const lookup = full.split('_')[1];
    expect(sql(`SELECT count(*) FROM devkit_api_key WHERE lookup='${lookup}' AND length(secret_hash)=32 AND encode(secret_hash,'hex') NOT LIKE '%${lookup}%'`)).toBe('1');
    const row = root.locator('[data-devkit-list="api-keys"] tbody tr').filter({ hasText: 'e2e editor ' + k.slice(-6) });
    await expect(row).toContainText('nfk_' + lookup);
    await expect(row).toContainText('فعال');

    const who = await keyCall('/v2/devkit/key?scope=schema:read', full);
    expect(who.status).toBe(200);
    expect(who.data).toMatchObject({ name: 'e2e editor ' + k.slice(-6), scopes: ['schema:read'] });
    const schema = await keyCall('/v2/devkit/rpc-schema', full);
    expect(schema.status).toBe(200);
    expect(Array.isArray(schema.data.rpcs)).toBe(true);
    const denied = await keyCall('/v2/devkit/key?scope=' + SCOPE, full);
    expect(denied.status).toBe(403);
    expect(denied.data).toMatchObject({ code: 'API_KEY_SCOPE_DENIED', meta: { scope: SCOPE } });
    expect(denied.data.message).toContain(SCOPE);
    expect((await keyCall('/v2/devkit/key')).data.code).toBe('API_KEY_REQUIRED');
    expect((await keyCall('/v2/devkit/key', full.slice(0, -2) + 'xx')).data.code).toBe('API_KEY_INVALID');
    await expect.poll(() => sql(`SELECT last_used_at IS NOT NULL FROM devkit_api_key WHERE lookup='${lookup}'`)).toBe('t');

    await page.reload();
    await expect(row).not.toContainText('هرگز');
    await row.locator('[data-devkit-action="revoke-key"]').click();
    await page.locator('[data-devkit-confirm="ok"]').click();
    await expect(row).toContainText('باطل‌شده');
    const revoked = await keyCall('/v2/devkit/key', full);
    expect(revoked.status).toBe(401);
    expect(revoked.data).toMatchObject({ code: 'API_KEY_REVOKED', message: 'این کلید API باطل شده است.' });
    await row.locator('[data-devkit-action="delete-key"]').click();
    await page.locator('[data-devkit-confirm="ok"]').click();
    await expect(row).toHaveCount(0);
  });

  test('key policies: max lifetime forces an expiry, wildcard grants, catalog-only scopes', async () => {
    const cur = (await dk.get('/settings')).settings;
    const { updatedBy, updatedAt, ...base } = cur;
    await dk.put('/settings', { ...base, apiKeys: { ...base.apiKeys, maxExpiryDays: 30, defaultExpiryDays: 0 } });
    const noExp = await consoleApi('POST', DK + '/api-keys', { name: 'e2e max ' + k.slice(-6), scopes: ['schema:read'], neverExpires: true });
    expect(noExp.status).toBe(400);
    expect(noExp.data.error).toBe('EXPIRY_REQUIRED');
    const tooFar = await consoleApi('POST', DK + '/api-keys', { name: 'e2e max ' + k.slice(-6), scopes: ['schema:read'], expiresInDays: 90 });
    expect(tooFar.data.error).toBe('EXPIRY_TOO_FAR');
    const unknown = await consoleApi('POST', DK + '/api-keys', { name: 'e2e max ' + k.slice(-6), scopes: ['nothere:write'], expiresInDays: 5 });
    expect(unknown.data.error).toBe('SCOPE_NOT_IN_CATALOG');
    const wild = await dk.post('/api-keys', { name: 'e2e wild ' + k.slice(-6), scopes: [SCOPE.split(':')[0] + ':*'], expiresInDays: 5 });
    expect(wild.item.expiresAt).toBeTruthy();
    expect((await keyCall('/v2/devkit/key?scope=' + SCOPE, wild.key)).status).toBe(200);
    await dk.put('/settings', { ...base, apiKeys: { ...base.apiKeys, allowWildcards: false } });
    await expect.poll(async () => (await keyCall('/v2/devkit/key?scope=' + SCOPE, wild.key)).status, { timeout: 20_000 }).toBe(403);
    const inUse = await consoleApi('DELETE', DK + '/api-scopes/' + SCOPE);
    expect(inUse.status === 200 || inUse.data.error === 'SCOPE_IN_USE').toBe(true);
    await dk.post('/api-keys/' + wild.item.id + '/revoke', {});
    await dk.del('/api-keys/' + wild.item.id);
    await dk.put('/settings', base);
  });
});

test.describe('devkit: rpc schema', () => {
  test('page opens with the fork-generated list; access setting gates the RPC', async ({ page }) => {
    await openDevkit(page, 'rpc-schema');
    const root = devkitRoot(page);
    await expect(root.locator('select[name="module"]')).toBeAttached();
    const box = root.locator('[data-devkit-settings="schema"]');
    await box.locator('[data-devkit-settings-toggle]').click();
    await box.locator('select[name="access"]').selectOption('console');
    await box.locator('[data-devkit-settings-save]').click();
    await expect.poll(async () => {
      const r = await rpc(P, 'devkit_rpc_schema', {});
      return r.status === 200 ? 'open' : errEnvelope(r).code;
    }, { timeout: 20_000 }).toBe('DEVKIT_SCHEMA_DISABLED');
    await box.locator('select[name="access"]').selectOption('session');
    await box.locator('[data-devkit-settings-save]').click();
    await expect.poll(async () => (await rpc(P, 'devkit_rpc_schema', {})).status, { timeout: 20_000 }).toBe(200);
  });

  test('english labels, and the pages fit 390px in dark mode without sideways scroll', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, locale: 'en', mode: 'dark', mobile: true });
    for (const [sub, title] of [['api-keys', 'API keys'], ['errors', 'Error messages'], ['rpc-schema', 'RPC schema']]) {
      await page.goto(CONSOLE_URL + '/#/devkit/' + sub);
      await expect(page.locator(`#devkit-root [data-devkit-page="${sub}"] h1`)).toHaveText(title);
      await expect(page.locator('#devkit-root [data-devkit-list]').first()).not.toBeEmpty();
      await noHorizontalScroll(page);
    }
  });
});

test.describe('devkit: no Unity package registry', () => {
  test('three pages in the sidebar and page search, the old address opens the error messages, registry and publish routes are gone', async ({ page }) => {
    await openDevkit(page, 'errors');
    const nav = page.locator('[data-devkit-group] a[data-devkit-nav]');
    if (await nav.count()) {
      expect(await nav.evaluateAll((els) => els.map((e) => e.getAttribute('data-devkit-nav')))).toEqual(['errors', 'api-keys', 'rpc-schema']);
    }
    await expect(page.locator('a[href="#/devkit/packages"]')).toHaveCount(0);
    expect(await page.evaluate(() => (window as any).NakamaDevkit.pages().map((p: any) => p.id))).toEqual(['errors', 'api-keys', 'rpc-schema']);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('پکیج');
    await expect(page.locator('.pl-pal-row[href="#/system/sdk"]')).toBeVisible();
    await expect(page.locator('.pl-pal-row[href^="#/devkit"]')).toHaveCount(0);
    await page.keyboard.press('Escape');

    await page.goto(CONSOLE_URL + '/#/devkit/packages');
    await expect(devkitRoot(page).locator('[data-devkit-page="errors"]')).toBeVisible();
    await expect(devkitRoot(page).locator('[data-devkit-settings="packages"]')).toHaveCount(0);

    const settings = (await dk.get('/settings')).settings;
    expect(settings.packages).toBeUndefined();
    expect((await consoleApi('GET', DK + '/packages')).status).toBe(404);
    const scopes = await dk.get('/api-scopes');
    expect(JSON.stringify(scopes)).not.toContain('upm:');
    for (const p of ['/upm/-/v1/search', '/upm/com.example.sdk', '/v2/devkit/packages']) {
      expect((await fetch(API_URL + p)).status, p).toBe(404);
    }
    expect((await fetch(API_URL + '/v2/devkit/packages', { method: 'POST', body: '{}' })).status).toBe(404);
  });
});
