import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, panel, root } from '../../lib/console';
import { base, deleteUser, devicePlayer, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { API_URL } from '../../lib/env';

const VERSION = 'e2e-' + Date.now().toString(36);
const BUNDLE = Buffer.alloc(1_600_000, 'roadstorm');
const CATALOG = Buffer.from(JSON.stringify({ m_LocatorId: 'AddressablesMainContentCatalog', e2e: VERSION }));

function errCode(res: { data: any }) {
  try { return JSON.parse(res.data.message).code; } catch { return res.data && res.data.message; }
}

test.describe.configure({ mode: 'serial' });

test.describe('downloadable content: upload from the console, catalog for the game', () => {
  const players: Player[] = [];
  let packId = '';

  test.afterAll(async () => {
    const list = await base.get('/content/packs');
    for (const p of list.items.filter((x: any) => String(x.version).startsWith('e2e'))) {
      if (p.active) await base.post('/content/packs/' + p.id + '/deactivate', {});
      await base.del('/content/packs/' + p.id);
    }
    sql(`DELETE FROM content_settings`);
    for (const p of players) await deleteUser(p.id).catch(() => {});
  });

  test('a build uploaded from the page becomes the active pack with public URLs', async ({ page }) => {
    await open(page, '#/system/content');
    await expect(root(page).locator('h1')).toHaveText('محتوای دانلودی');
    await root(page).locator('[data-ct-new]').click();
    const form = panel(page).locator('[data-ct-upload]');
    await expect(form).toBeVisible();
    await form.locator('[data-ct-field="platform"]').selectOption('Android');
    await form.locator('[data-ct-field="version"]').fill(VERSION);
    await form.locator('[data-ct-field="min"]').fill('900.0');
    await form.locator('[data-ct-input="files"]').setInputFiles([
      { name: 'catalog_e2e.json', mimeType: 'application/json', buffer: CATALOG },
      { name: 'catalog_e2e.hash', mimeType: 'text/plain', buffer: Buffer.from('e2ehash' + VERSION) },
      { name: 'cars_assets_all.bundle', mimeType: 'application/octet-stream', buffer: BUNDLE },
    ]);
    await expect(form.locator('[data-ct-summary]')).toContainText('۳ فایل');
    await expect(form.locator('[data-ct-summary]')).toContainText('catalog_e2e.json');
    await panel(page).locator('[data-ct-send]').click();
    const detail = panel(page).locator('[data-ct-pack]');
    await expect(detail).toBeVisible({ timeout: 60_000 });
    await expect(detail.locator('[data-ct-status="ready"]')).toContainText('فعال');
    packId = (await detail.getAttribute('data-ct-pack')) || '';
    expect(packId).toMatch(/^[0-9a-f-]{36}$/);
    await expect(detail.locator('[data-ct-url="catalog"]')).toContainText('/v2/content/files/' + packId + '/catalog_e2e.json');
    await expect(detail.locator('[data-ct-files]')).toContainText('cars_assets_all.bundle');
    const got = await base.get('/content/packs/' + packId);
    expect(got.item).toMatchObject({ platform: 'Android', version: VERSION, status: 'ready', active: true, catalogHash: 'e2ehash' + VERSION, fileCount: 3, minAppVersion: '900.0' });
  });

  test('a device player fetches the catalog for its build and downloads the files', async () => {
    const p = await devicePlayer('e2e-ct-' + Date.now());
    players.push(p);
    const res = await rpc(p, 'content_catalog', { platform: 'Android', appVersion: '900.2.1' });
    expect(res.status).toBe(200);
    expect(res.data.pack).toMatchObject({ id: packId, version: VERSION, catalogHash: 'e2ehash' + VERSION });
    const cat = await fetch(res.data.pack.catalogUrl);
    expect(cat.status).toBe(200);
    expect(Buffer.from(await cat.arrayBuffer()).equals(CATALOG)).toBe(true);
    const bundle = await fetch(res.data.pack.baseUrl + '/cars_assets_all.bundle');
    expect(bundle.status).toBe(200);
    expect(bundle.headers.get('cache-control')).toContain('immutable');
    expect(Buffer.from(await bundle.arrayBuffer()).equals(BUNDLE)).toBe(true);
    const old = await rpc(p, 'content_catalog', { platform: 'Android', appVersion: '899.9' });
    expect(old.data.pack === null || old.data.pack.id !== packId).toBe(true);
    expect(errCode(await rpc(p, 'content_catalog', {}))).toBe('PLATFORM_REQUIRED');
    const pub = await fetch(API_URL + '/v2/content/catalog?platform=Android&appVersion=900.5');
    expect(pub.status).toBe(200);
    expect((await pub.json()).pack.id).toBe(packId);
  });

  test('the preview, deactivate and delete from the page', async ({ page }) => {
    await open(page, '#/system/content');
    await root(page).locator('[data-ct-preview="version"]').fill('۹۰۰.۳');
    await root(page).locator('[data-ct-preview="go"]').click();
    await expect(root(page).locator('[data-ct-preview-result]')).toContainText(VERSION);
    await root(page).locator('[data-ct-row="Android:' + VERSION + '"]').click();
    await panel(page).locator('[data-ct-action="deactivate"]').click();
    await expect(panel(page).locator('[data-ct-action="activate"]')).toBeVisible();
    const p = players[0];
    const res = await rpc(p, 'content_catalog', { platform: 'Android', appVersion: '900.2.1' });
    expect(res.data.pack === null || res.data.pack.id !== packId).toBe(true);
    await panel(page).locator('[data-ct-action="delete"]').click();
    await page.locator('[data-economy="dialog"] button').last().click();
    await expect(panel(page)).toHaveCount(0);
    expect((await base.del('/content/packs/' + packId)).status).toBe(404);
  });

  test('settings save a policy, Ctrl K finds the page', async ({ page }) => {
    await open(page, '#/system/content/settings');
    const form = root(page).locator('[data-settings-form="content"]');
    await expect(form).toBeVisible();
    await form.getByText('نسخهٔ برنامه لازم باشد').click();
    await form.locator('[data-settings-save]').click();
    await expect.poll(async () => (await base.get('/content/settings')).settings.packs.requireAppVersion).toBe(true);
    const p = players[0];
    await expect.poll(async () => errCode(await rpc(p, 'content_catalog', { platform: 'Android' })), { timeout: 15_000 }).toBe('APP_VERSION_REQUIRED');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('محتوای دانلودی');
    await expect(page.locator('.pl-pal-row').first().locator('.pl-pal-label')).toHaveText('محتوای دانلودی');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/system\/content(\/settings)?$/);
    await expect(root(page).locator('h1')).toHaveText('محتوای دانلودی');
  });

  test('English, dark, 390px', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/system/content');
    await expect(root(m).locator('h1')).toHaveText('Downloadable content');
    await expect(root(m).locator('[data-ct-live]')).toContainText('Live platforms');
    await noHorizontalScroll(m);
    await open(m, '#/system/content/settings');
    await expect(root(m).locator('[data-settings-form="content"]')).toContainText('Part size (KB)');
    await noHorizontalScroll(m);
  });
});
