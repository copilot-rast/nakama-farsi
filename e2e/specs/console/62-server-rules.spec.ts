import { expect, test } from '../../lib/fixtures';
import { fillTranslated, noHorizontalScroll, open, root } from '../../lib/console';
import { base, deleteUser, devicePlayer, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { API_URL } from '../../lib/env';

const DEFAULTS = {
  maintenance: { enabled: false, message: {}, until: null, allowRpcs: [], allowTagIds: [] },
  storage: { blockWrites: false, blockDeletes: false, allowCollections: [] },
  accountDelete: { enabled: true, mode: 'grace', graceDays: 14 },
};

function errCode(res: { data: any }) {
  try { return JSON.parse(res.data.message).code; } catch { return res.data && res.data.message; }
}

async function storageWrite(p: Player, collection: string) {
  const r = await fetch(API_URL + '/v2/storage', {
    method: 'PUT',
    headers: { Authorization: 'Bearer ' + p.token, 'Content-Type': 'application/json' },
    body: JSON.stringify({ objects: [{ collection, key: 'k', value: '{"a":1}' }] }),
  });
  return { status: r.status, data: await r.json().catch(() => null) };
}

test.describe.configure({ mode: 'serial' });

test.describe('server rules: maintenance, storage lock, account deletion', () => {
  const players: Player[] = [];

  test.afterAll(async () => {
    await base.put('/server-rules', DEFAULTS);
    sql(`DELETE FROM base_server_rules`);
    for (const p of players) await deleteUser(p.id).catch(() => {});
  });

  test('the page saves maintenance; players are refused with the message, the version call still answers', async ({ page }) => {
    const p = await devicePlayer('e2e-sr-' + Date.now());
    players.push(p);
    expect((await rpc(p, 'economy_catalog', { lang: 'fa' })).status).toBe(200);
    await open(page, '#/system/server-rules');
    const form = root(page).locator('[data-settings-form="server-rules"]');
    await expect(form).toBeVisible();
    await form.getByText('حالت تعمیر روشن باشد').click();
    await fillTranslated(form, 'maintenance.message', { fa: 'تا ساعت ۵ تعمیر داریم' });
    await form.locator('[data-settings-save]').click();
    await expect(root(page).locator('.eco-notice').filter({ hasText: 'حالت تعمیر روشن است' })).toBeVisible();
    await expect.poll(async () => (await base.get('/server-rules')).maintenance.enabled).toBe(true);
    await expect.poll(async () => errCode(await rpc(p, 'economy_catalog', { lang: 'fa' })), { timeout: 15_000 }).toBe('MAINTENANCE');
    const refused = await rpc(p, 'economy_catalog', { lang: 'fa' });
    expect(JSON.parse(refused.data.message).message).toBe('تا ساعت ۵ تعمیر داریم');
    const ver = await rpc(p, 'base_app_version', { platform: 'android', version: '1.0.0', lang: 'fa' });
    expect(ver.status).toBe(200);
    expect(ver.data.maintenance).toMatchObject({ on: true, message: 'تا ساعت ۵ تعمیر داریم' });
    expect((await rpc(p, 'base_server_status', {})).data.maintenance.on).toBe(true);
    await base.put('/server-rules', DEFAULTS);
    await expect.poll(async () => (await rpc(p, 'economy_catalog', { lang: 'fa' })).status, { timeout: 15_000 }).toBe(200);
  });

  test('storage lock refuses game writes except open collections', async () => {
    const p = await devicePlayer('e2e-sr-st-' + Date.now());
    players.push(p);
    expect((await storageWrite(p, 'saves')).status).toBe(200);
    await base.put('/server-rules', { ...DEFAULTS, storage: { blockWrites: true, blockDeletes: true, allowCollections: ['settings'] } });
    await expect.poll(async () => (await storageWrite(p, 'saves')).status, { timeout: 15_000 }).toBe(403);
    expect(JSON.stringify((await storageWrite(p, 'saves')).data)).toContain('STORAGE_WRITE_BLOCKED');
    expect((await storageWrite(p, 'settings')).status).toBe(200);
    await base.put('/server-rules', DEFAULTS);
  });

  test('account deletion with a grace period, cancel, and the console count', async ({ page }) => {
    const p = await devicePlayer('e2e-sr-del-' + Date.now());
    players.push(p);
    expect(errCode(await rpc(p, 'base_account_delete', {}))).toBe('CONFIRM_REQUIRED');
    const del = await rpc(p, 'base_account_delete', { confirm: true });
    expect(del.data).toMatchObject({ deleted: false, graceDays: 14 });
    expect((await base.get('/server-rules')).pendingDeletions).toBeGreaterThan(0);
    await open(page, '#/system/server-rules');
    await expect(root(page).locator('.eco-notice').filter({ hasText: 'در صف حذف است' })).toBeVisible();
    expect((await rpc(p, 'base_account_delete_cancel', {})).status).toBe(200);
    expect(errCode(await rpc(p, 'base_account_delete_cancel', {}))).toBe('NO_PENDING_DELETION');
  });

  test('English, dark, 390px', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/system/server-rules');
    await expect(root(m).locator('h1')).toHaveText('Server rules');
    await expect(root(m).locator('[data-settings-form="server-rules"]')).toContainText('Maintenance is on');
    await noHorizontalScroll(m);
  });
});
