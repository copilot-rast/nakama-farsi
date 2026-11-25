import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';

const RUN = Date.now().toString(36);
const GAME = 'e2e' + RUN.slice(-6);
const ROOT = '#ins-root';

async function openIns(page: Page, hash: string) {
  await page.goto(CONSOLE_URL + '/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.locator(ROOT + ' [data-ins-page]').waitFor({ timeout: 15_000 });
  await expect(page.locator(ROOT).getByText(/در حال بارگذاری|Loading/)).toHaveCount(0, { timeout: 15_000 });
}

test.describe.configure({ mode: 'serial' });

test.describe('server-filled tags: settings page, defaults, page search', () => {
  const players: Player[] = [];

  test.afterAll(async () => {
    sql(`DELETE FROM insights_settings`);
    sql(`DELETE FROM insights_event WHERE game='${GAME}'`);
    sql(`DELETE FROM insights_event_user WHERE game='${GAME}'`);
    sql(`DELETE FROM insights_event_user_day WHERE game='${GAME}'`);
    for (const p of players) await deleteUser(p.id).catch(() => {});
  });

  test('defaults, edit, validate, save and reset', async ({ page }) => {
    sql(`DELETE FROM insights_settings`);
    const d = await base.get('/insights/events/settings');
    expect(d.context).toMatchObject({ enabled: true, conflict: 'override', flagMismatch: true, keepClientOutside: true, graceSeconds: 10 });
    expect(d.context.fills).toEqual([
      { group: 'mission', source: 'campaign.stage', enabled: true },
      { group: 'car', source: 'campaign.item', enabled: true },
      { group: 'chapter', source: 'campaign.chapter', enabled: false },
    ]);
    expect(d.sources.map((s: any) => s.key)).toEqual(expect.arrayContaining(['campaign.stage', 'campaign.item', 'campaign.chapter', 'campaign.attempt']));

    await openIns(page, '#/insights/context');
    await expect(page.locator(ROOT + ' h1')).toHaveText('تگ‌هایی که سرور پر می‌کند');
    await expect(page.locator('[data-ins-fill]')).toHaveCount(3);
    await expect(page.locator('select[name="conflict"]')).toHaveValue('override');
    await expect(page.locator('[data-ins-fill="chapter"] [name="fillEnabled"]')).not.toBeChecked();
    await page.locator('select[name="conflict"]').selectOption('reject');
    await page.locator('[data-ins-fill="chapter"] [name="fillEnabled"]').check();
    await page.locator('[name="keepClientOutside"]').uncheck();
    await page.locator('input[name="graceSeconds"]').fill('۳۰');
    await page.locator('[data-ins-fill-add]').click();
    const added = page.locator('[data-ins-fill]').last();
    await added.locator('input[name="group"]').fill('Bad Group');
    await page.locator('[data-ins-context-save]').click();
    await expect(page.locator('[data-ins-msg]')).toContainText('گروه تگ نامعتبر است');
    await added.locator('input[name="group"]').fill('attempt');
    await added.locator('select[name="source"]').selectOption('campaign.attempt');
    await page.locator('[data-ins-context-save]').click();
    await expect.poll(async () => (await base.get('/insights/events/settings')).context.conflict).toBe('reject');
    const saved = (await base.get('/insights/events/settings')).context;
    expect(saved).toMatchObject({ conflict: 'reject', keepClientOutside: false, graceSeconds: 30 });
    expect(saved.fills).toHaveLength(4);
    expect(saved.fills[2]).toEqual({ group: 'chapter', source: 'campaign.chapter', enabled: true });
    expect(saved.fills[3]).toEqual({ group: 'attempt', source: 'campaign.attempt', enabled: true });

    await page.reload();
    await openIns(page, '#/insights/context');
    await expect(page.locator('[data-ins-fill]')).toHaveCount(4);
    await page.locator('[data-ins-fill="attempt"] [data-ins-fill-delete]').click();
    await page.locator('[data-ins-context-reset]').click();
    await expect(page.locator('[data-ins-fill]')).toHaveCount(3);
    await page.locator('[data-ins-context-save]').click();
    await expect.poll(async () => (await base.get('/insights/events/settings')).context.conflict).toBe('override');
    expect((await base.get('/insights/events/settings')).context).toMatchObject({ keepClientOutside: true, graceSeconds: 10 });
  });

  test('a player with no open stage keeps the tags it sends', async () => {
    const p = await devicePlayer(`e2e-st-${RUN}-device`);
    players.push(p);
    const r = await rpc(p, 'insights_send', { batchId: `st-${RUN}-0001`, game: GAME, events: [{ name: 'ShopView', tags: [{ tagGroup: 'car', id: 'tiggo8' }], value: 1 }, { name: 'DistanceOverall', value: 2 }] });
    expect(r.status).toBe(200);
    expect(r.data.accepted).toBe(2);
    await expect.poll(async () => (await base.get('/insights/outbox')).pending, { timeout: 20_000 }).toBe(0);
    const rows = sql(`SELECT name || ' ' || tags::text FROM insights_event WHERE game='${GAME}' ORDER BY name`).trim().split('\n');
    expect(rows).toEqual(['DistanceOverall {}', 'ShopView {"car": "tiggo8"}']);
  });

  test('page search, English and a phone', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await openIns(page, '#/insights/context');
    await expect(page.locator('[data-ins-fill]')).toHaveCount(3);
    const scroll = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    expect(scroll).toBeLessThanOrEqual(1);
    await page.setViewportSize({ width: 1440, height: 900 });
    await openIns(page, '#/insights/overview');
    expect(await page.evaluate(() => (window as any).NakamaInsights.pages().map((p: any) => p.id))).toContain('insights/context');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('سرور پر');
    await expect(page.locator('.pl-pal-row[href="#/insights/context"]')).toBeVisible();
    await page.locator('.pl-pal-row[href="#/insights/context"]').click();
    await expect(page.locator(ROOT + ' h1')).toHaveText('تگ‌هایی که سرور پر می‌کند');
    await page.locator('.pl-seg-lang [data-locale="en"]').first().click();
    await expect(page.locator(ROOT + ' h1')).toHaveText('Server-filled tags');
    await expect(page.locator('input[name="graceSeconds"]')).toHaveValue('10');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').first().click();
    await expect(page.locator('input[name="graceSeconds"]')).toHaveValue('۱۰');
  });
});
