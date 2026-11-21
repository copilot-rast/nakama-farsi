import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';

const RUN = Date.now().toString(36);
const GAME = 'e2e' + RUN.slice(-6);
const ROOT = '#ins-root';

function daysAgo(n: number) {
  return new Date(Date.now() - n * 86400000).toISOString();
}

function day(n: number) {
  const d = new Date(Date.now() - n * 86400000);
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

async function openIns(page: Page, hash: string) {
  await page.goto(CONSOLE_URL + '/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.locator(ROOT + ' [data-ins-page]').waitFor({ timeout: 15_000 });
  await expect(page.locator(ROOT).getByText(/در حال بارگذاری|Loading/)).toHaveCount(0, { timeout: 15_000 });
}

function q(extra = '') {
  return `?game=${GAME}&from=${day(3)}&to=${day(0)}${extra}`;
}

async function waitDelivered() {
  await expect.poll(async () => (await base.get('/insights/outbox')).pending, { timeout: 20_000 }).toBe(0);
}

test.describe.configure({ mode: 'serial' });

test.describe('game analytics: event standard, dashboards and settings', () => {
  const players: Player[] = [];
  let funnelId = '';

  test.afterAll(async () => {
    if (funnelId) await base.del('/insights/funnels/' + funnelId);
    await base.del(`/insights/events/catalog/groups/car?game=${GAME}`);
    await base.del(`/insights/events/catalog/groups/mission?game=${GAME}`);
    await base.del(`/insights/events/catalog/events/DistanceOverall?game=${GAME}`);
    sql(`DELETE FROM insights_settings`);
    sql(`DELETE FROM insights_event WHERE game='${GAME}'`);
    sql(`DELETE FROM insights_event_user WHERE game='${GAME}'`);
    sql(`DELETE FROM insights_event_user_day WHERE game='${GAME}'`);
    for (const p of players) await deleteUser(p.id).catch(() => {});
  });

  test('catalog from the console, events from four players, owner sketch aggregations', async ({ page }) => {
    await openIns(page, '#/insights/catalog?game=' + GAME);
    await expect(page.locator(ROOT + ' h1')).toHaveText('فهرست رویدادها و تگ‌ها');
    await page.locator('[data-ins-group-new]').click();
    const panel = page.locator('[data-ins-panel="group"]');
    await panel.locator('input[name="key"]').fill('car');
    await panel.locator('input[name="nameFa"]').fill('ماشین');
    await panel.locator('[name="closed"]').check();
    await panel.locator('textarea[name="values"]').fill('345345,سانتافه,SantaFe\n111,پراید,Pride');
    await panel.locator('[data-ins-group-save]').click();
    await expect(page.locator(ROOT + ' [data-ins-table="groups"]')).toContainText('car');
    await expect(page.locator(ROOT + ' [data-ins-table="groups"]')).toContainText('۲ مقدار');
    await base.put(`/insights/events/catalog/groups/mission?game=${GAME}`, { nameEn: 'Mission', enabled: true });
    await page.locator('[data-ins-event-new]').click();
    const ev = page.locator('[data-ins-panel="event"]');
    await ev.locator('input[name="name"]').fill('DistanceOverall');
    await ev.locator('input[name="pathGroups"]').fill('car, mission');
    await ev.locator('input[name="max"]').fill('۱۰۰۰۰۰');
    await ev.locator('[data-ins-event-save]').click();
    await expect(page.locator(ROOT + ' [data-ins-table="events"]')).toContainText('car, mission');

    for (let i = 0; i < 4; i++) {
      const p = await devicePlayer(`e2e-ga-${RUN}-${i}-device`);
      players.push(p);
      const sid = `${GAME}-s${i}`;
      expect((await rpc(p, 'insights_session', { event: 'start', sessionId: sid, platform: i % 2 ? 'ios' : 'android', appVersion: '1.0.' + i, game: GAME })).status).toBe(200);
      const events: any[] = [
        { name: 'LevelStart', tags: [{ tagGroup: 'mission', id: '3434343' }], time: daysAgo(i === 0 ? 1 : 0) },
        { name: 'DistanceOverall', tags: [{ tagGroup: 'mission', id: '3434343' }, { tagGroup: 'car', id: '345345' }], value: 150, time: daysAgo(i === 0 ? 1 : 0) },
        { path: `DistanceOverall:111:m${i}`, value: 10 },
      ];
      if (i < 2) events.push({ name: 'LevelEnd', tags: { mission: '3434343' } });
      if (i === 3) events.push({ name: 'DistanceOverall', tags: { car: '999' }, value: 5 }, { name: '1bad' });
      const res = await rpc(p, 'insights_send', { batchId: `${GAME}-b${i}-0001`, sessionId: sid, platform: i % 2 ? 'ios' : 'android', appVersion: '1.0.' + i, game: GAME, events });
      expect(res.status).toBe(200);
      if (i === 3) expect(res.data).toMatchObject({ accepted: 3, flagged: 1, rejected: 1 });
      const dup = await rpc(p, 'insights_send', { batchId: `${GAME}-b${i}-0001`, game: GAME, events });
      expect(dup.data.duplicate).toBe(true);
      expect((await rpc(p, 'insights_session', { event: 'end', sessionId: sid, seconds: 90 })).data.status).toBe('ended');
    }
    await rpc(players[0], 'insights_session', { event: 'start', sessionId: `${GAME}-s0b`, game: GAME });
    await rpc(players[0], 'insights_send', { batchId: `${GAME}-b0-0002`, sessionId: `${GAME}-s0b`, game: GAME, events: [{ name: 'LevelEnd', tags: { mission: '3434343' } }] });
    await waitDelivered();
    const total = await base.get('/insights/explore' + q('&name=DistanceOverall&metric=sum'));
    expect(total.rows[0].value).toBe(640);
    const combo = await base.get('/insights/explore' + q('&name=DistanceOverall&metric=sum&groupBy=tag:car,tag:mission,user'));
    expect(combo.rows.filter((r: any) => r.keys[0] === '345345' && r.keys[1] === '3434343')).toHaveLength(4);
  });

  test('overview, active users and retention pages', async ({ page }) => {
    await openIns(page, '#/insights/overview' + q());
    await expect(page.locator(ROOT + ' h1')).toHaveText('نمای کلی');
    await expect(page.locator(ROOT + ' [data-ins-kpi="new"]')).toContainText('۴');
    await expect(page.locator(ROOT + ' [data-ins-kpi="dau"]')).toContainText('۴');
    await expect(page.locator(ROOT + ' [data-ins-chart="overview-dau"] svg path.ia-line')).toHaveCount(2);
    await expect(page.locator(ROOT + ' [data-ins-chart="top-events"]')).toContainText('DistanceOverall');
    await expect(page.locator('.x-drawer [data-ins-group="insights"]')).toContainText('آنالیتیکس بازی');
    await page.locator('a[href="#/insights/active"]:visible').first().click();
    await expect(page.locator(ROOT + ' h1')).toHaveText('کاربران فعال');
    await expect(page.locator(ROOT + ' [data-ins-table="active"]')).toBeVisible();
    const download = page.waitForEvent('download');
    await page.locator('[data-ins-csv="active-users"]').click();
    expect((await download).suggestedFilename()).toBe('active-users.csv');
    await openIns(page, '#/insights/retention' + q('&days=1'));
    await expect(page.locator(ROOT + ' [data-ins-kpi="d1"]')).toContainText('۱۰۰٫۰٪');
    await expect(page.locator(ROOT + ' [data-ins-table="cohorts"]')).toContainText('۱۰۰٫۰٪');
  });

  test('explorer groups by tags and the funnel shows drop-off', async ({ page }) => {
    await openIns(page, '#/insights/explore' + q('&name=DistanceOverall&metric=sum&groupBy=tag:car'));
    await expect(page.locator(ROOT + ' [data-ins-table="explore"]')).toContainText('345345');
    await expect(page.locator(ROOT + ' [data-ins-table="explore"]')).toContainText('۶۰۰');
    await page.locator('select[name="bucket"]').selectOption('day');
    await page.locator('[data-ins-run]').click();
    await expect(page.locator(ROOT + ' [data-ins-chart="explore"] svg')).toBeVisible();
    await openIns(page, '#/insights/funnels' + q());
    await page.locator('[data-ins-funnel-new]').click();
    const fp = page.locator('[data-ins-panel="funnel"]');
    await fp.locator('input[name="name"]').fill('Level ' + RUN);
    await fp.locator('input[name="game"]').fill(GAME);
    await fp.locator('input[name="window"]').fill('60');
    const steps = fp.locator('[data-ins-step]');
    await steps.nth(0).locator('input[name="step-name"]').fill('LevelStart');
    await steps.nth(0).locator('input[name="step-tags"]').fill('mission=3434343');
    await steps.nth(1).locator('input[name="step-name"]').fill('LevelEnd');
    await fp.locator('[data-ins-funnel-save]').click();
    await expect(page.locator(ROOT + ' [data-ins-step-result="1"]')).toContainText('۲۵٫۰٪');
    funnelId = (await base.get('/insights/funnels')).items.find((f: any) => f.name === 'Level ' + RUN).id;
    await expect(page.locator(ROOT + ' [data-ins-step-result="0"]')).toContainText('۴');
  });

  test('session list opens a timeline of the session events', async ({ page }) => {
    await openIns(page, '#/insights/sessions?userId=' + players[1].id);
    await page.locator(ROOT + ' [data-ins-table="sessions"] tbody tr').first().click();
    const tl = page.locator('[data-ins-panel="timeline"] [data-ins-timeline]');
    await expect(tl.locator('[data-ins-event="session_start"]')).toHaveCount(1);
    await expect(tl.locator('[data-ins-event="LevelEnd"]')).toHaveCount(1);
    await expect(tl.locator('[data-ins-event="session_end"]')).toHaveCount(1);
  });

  test('storage shows the queue, settings save and validate, imports list the kinds', async ({ page }) => {
    await openIns(page, '#/insights/storage');
    await expect(page.locator('select[name="driver"] option')).toHaveCount(4);
    await expect(page.locator(ROOT + ' [data-ins-section="outbox"]')).toContainText('صف خالی است');
    await page.locator('select[name="driver"]').selectOption('clickhouse');
    await page.locator('input[name="url"]').fill('http://127.0.0.1:1');
    await page.locator('[data-ins-storage-test]').click();
    await expect(page.locator('[data-ins-msg]')).toContainText('اتصال برقرار نشد');
    await openIns(page, '#/insights/settings');
    const maxTags = page.locator('input[name="maxTags"]');
    await maxTags.fill('۴۰');
    await page.locator('[data-ins-settings-save]').click();
    await expect(page.locator('[data-ins-msg]')).toContainText('events.maxTags');
    await maxTags.fill('۶');
    await page.locator('[data-ins-settings-save]').click();
    await expect.poll(async () => (await base.get('/insights/events/settings')).events.maxTags).toBe(6);
    await openIns(page, '#/insights/imports');
    await page.locator('[data-ins-connector-new]').click();
    await expect(page.locator('[data-ins-panel="connector"] select[name="kind"] option')).toHaveCount(3);
    await page.locator('[data-ins-panel="connector"] select[name="kind"]').selectOption('gameanalytics');
    await expect(page.locator('[data-ins-panel="connector"] textarea[name="config"]')).toHaveValue(/bucket/);
  });

  test('English uses Latin digits and the page fits a phone', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await openIns(page, '#/insights/overview' + q());
    const scroll = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    expect(scroll).toBeLessThanOrEqual(1);
    await page.setViewportSize({ width: 1440, height: 900 });
    await openIns(page, '#/insights/overview' + q());
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(page.locator(ROOT + ' h1')).toHaveText('Overview');
    await expect(page.locator(ROOT + ' [data-ins-kpi="new"]')).toContainText('4');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator(ROOT + ' h1')).toHaveText('نمای کلی');
  });
});
