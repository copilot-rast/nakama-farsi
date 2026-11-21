import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { base, uid } from '../../lib/api';
import { noHorizontalScroll } from '../../lib/console';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const SOURCE = 'E2E.Card' + k;
let savedSettings: any = null;

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#itemkinds-root');
const panel = (page: Page) => page.locator('[data-ik=panel] [role=dialog]');

async function openPage(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/itemkinds/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-ik-page="itemkinds/${sub}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

test.beforeAll(async () => {
  savedSettings = (await base.get('/itemkinds/settings')).settings;
});

test.afterAll(async () => {
  const list = (await base.get('/itemkinds/routes')).routes || [];
  for (const r of list) if (r.sourceType === SOURCE) await base.del('/itemkinds/routes/' + r.id);
  if (savedSettings) await base.put('/itemkinds/settings', savedSettings);
});

test('kinds API answers and the sidebar group only shows when a kind is registered', async ({ page }) => {
  const d = await base.get('/itemkinds/kinds');
  expect(Array.isArray(d.kinds)).toBe(true);
  await openPage(page, 'routes');
  const group = page.locator('.x-drawer [data-ik-group="itemkinds"]');
  if (d.kinds.length === 0) await expect(group).toHaveCount(0);
  else await expect(group.locator('[data-ik-nav]')).toHaveCount(d.kinds.filter((x: any) => x.enabled !== false).length + 3);
});

test('admin adds, edits and removes a route from the router page', async ({ page }) => {
  await openPage(page, 'routes');
  await root(page).locator('[data-ik-route-new]').click();
  const p = panel(page);
  await expect(p).toBeVisible();
  await p.locator('input[name=sourceType]').fill(SOURCE);
  await p.locator('textarea[name=match]').fill('{"category.$.$name": "Vehicle"}');
  await p.locator('select[name=role]').selectOption('dependent');
  await p.locator('[data-ik-save=route]').click();
  await expect(p).toHaveCount(0);
  const row = root(page).locator(`[data-ik-row]`).filter({ hasText: SOURCE });
  await expect(row).toHaveCount(1);
  await expect(row).toContainText('وابسته');
  const stored = ((await base.get('/itemkinds/routes')).routes || []).find((r: any) => r.sourceType === SOURCE);
  expect(stored.match).toEqual({ 'category.$.$name': 'Vehicle' });
  expect(stored.origin).toBe('');

  await row.click();
  await expect(p).toBeVisible();
  await p.locator('textarea[name=match]').fill('{not json');
  await p.locator('[data-ik-save=route]').click();
  await expect(p.locator('[data-ik-msg]')).toContainText('JSON');
  await p.locator('textarea[name=match]').fill('');
  await p.locator('input[name=priority]').fill('۵۰');
  await p.locator('[data-ik-save=route]').click();
  await expect(p).toHaveCount(0);
  const edited = ((await base.get('/itemkinds/routes')).routes || []).find((r: any) => r.sourceType === SOURCE);
  expect(edited.priority).toBe(50);
  expect(edited.match).toEqual({});

  await root(page).locator(`[data-ik-row]`).filter({ hasText: SOURCE }).click();
  await p.locator(`[data-ik-route-delete="${edited.id}"]`).click();
  await page.locator('[data-ik=confirm] [data-ik-confirm=ok]').click();
  await expect(root(page).locator(`[data-ik-row]`).filter({ hasText: SOURCE })).toHaveCount(0);
});

test('settings page saves router policies and rejects bad values', async ({ page }) => {
  await openPage(page, 'settings');
  const r = root(page);
  await r.locator('select[name="routing.onDelete"]').selectOption('keep');
  await r.locator('input[name="routing.keepRuns"]').fill('۲۰');
  await r.locator('select[name="inventory.profile"]').selectOption('always');
  await r.locator('[data-ik-save="itemkinds/settings"]').click();
  await expect(page.locator('[data-ik=toasts]')).toContainText('ذخیره شد');
  const s = (await base.get('/itemkinds/settings')).settings;
  expect(s.routing.onDelete).toBe('keep');
  expect(s.routing.keepRuns).toBe(20);
  expect(s.inventory.profile).toBe('always');
  await r.locator('input[name="routing.maxDepth"]').fill('99');
  await r.locator('[data-ik-save="itemkinds/settings"]').click();
  await expect(r.locator('[data-ik-msg]')).toContainText('تنظیم درست نیست');
});

test('runs page and the router page fit a phone screen', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 800 });
  await openPage(page, 'runs');
  await noHorizontalScroll(page);
  await openPage(page, 'routes');
  await noHorizontalScroll(page);
});
