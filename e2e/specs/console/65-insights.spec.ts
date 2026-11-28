import zlib from 'node:zlib';
import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, panel, root } from '../../lib/console';
import { base, deleteUser, devicePlayer, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';

const TAG = 'e2e' + Date.now().toString(36);
const MAP = TAG + '_desert';

function errCode(res: { data: any }) {
  try { return JSON.parse(res.data.message).code; } catch { return res.data && res.data.message; }
}

function race(points: number) {
  const lines = ['t,x,z,player,kmh'];
  for (let i = 0; i < points; i++) {
    const car = i % 3;
    lines.push([(i * 0.5).toFixed(1), (Math.cos(i / 20) * 80 + car * 3).toFixed(2), (Math.sin(i / 20) * 50).toFixed(2), car ? 'bot' + car : 'me', (90 + (i % 40)).toString()].join(','));
  }
  return zlib.gzipSync(lines.join('\n')).toString('base64');
}

test.describe.configure({ mode: 'serial' });

test.describe('analytics: sessions, retention and heatmaps', () => {
  const players: Player[] = [];

  test.afterAll(async () => {
    await base.del('/insights/heatmaps/' + MAP);
    sql(`DELETE FROM insights_settings`);
    for (const p of players) await deleteUser(p.id).catch(() => {});
  });

  test('sessions from two device players show on the overview and the sessions list', async ({ page }) => {
    for (let i = 0; i < 2; i++) {
      const p = await devicePlayer('e2e-in-' + i + '-' + Date.now());
      players.push(p);
      const sid = TAG + '-s' + i;
      const start = await rpc(p, 'insights_session', { event: 'start', sessionId: sid, platform: TAG, appVersion: '1.4.0', device: 'SM-A515F', os: 'Android 14' });
      expect(start.status).toBe(200);
      expect(start.data).toMatchObject({ status: 'open', created: true, pingSeconds: 60 });
      expect((await rpc(p, 'insights_session', { event: 'ping', sessionId: sid })).data.status).toBe('open');
      const end = await rpc(p, 'insights_session', { event: 'end', sessionId: sid, seconds: 1 });
      expect(end.data).toMatchObject({ status: 'ended', seconds: 1 });
    }
    expect(errCode(await rpc(players[0], 'insights_session', { event: 'jump', sessionId: 'x' }))).toBe('ANALYTICS_INVALID_SESSION');
    const ov = await base.get('/insights/overview?platform=' + TAG);
    expect(ov.totals).toMatchObject({ sessions: 2, players: 2, newPlayers: 2 });
    expect(ov.active.dau).toBe(2);
    await open(page, '#/system/insights?platform=' + TAG);
    await expect(root(page).locator('h1')).toHaveText('آنالیتیکس');
    await expect(root(page).locator('[data-in-kpi="dau"]')).toContainText('۲');
    await expect(root(page).locator('[data-in-kpi="sessions"]')).toContainText('۲');
    await expect(root(page).locator('[data-in-chart="players"] svg')).toBeVisible();
    await expect(root(page).locator('[data-in-retention]')).toContainText('روز ۱');
    await open(page, '#/system/insights/sessions?userId=' + players[0].id);
    await expect(root(page).locator('[data-in-sessions]')).toContainText(TAG + '-s0');
    await expect(root(page).locator('[data-in-sessions]')).toContainText('تمام‌شده');
  });

  test('a race heatmap uploaded by the game shows on the map with its path', async ({ page }) => {
    const p = players[0];
    const up = await rpc(p, 'insights_heatmap', { map: MAP, uploadId: TAG + '-race1', format: 'csv.gz', data: race(240) });
    expect(up.status).toBe(200);
    expect(up.data).toMatchObject({ accepted: 240, players: 3, duplicate: false });
    expect((await rpc(p, 'insights_heatmap', { map: MAP, uploadId: TAG + '-race1', format: 'csv.gz', data: race(240) })).data.duplicate).toBe(true);
    expect((await rpc(p, 'insights_heatmap', { map: MAP, uploadId: TAG + '-race1', part: 1, points: [[0, 1, 1, 'me', 100], [0.5, 2, 1, 'me', 110]] })).data.accepted).toBe(2);
    expect(errCode(await rpc(p, 'insights_heatmap', { map: 'bad map!', uploadId: 'x', data: '1,2,3' }))).toBe('HEATMAP_INVALID');
    const cfg = await rpc(p, 'insights_config', {});
    expect(cfg.data.heatmap.maps).toContain(MAP);
    await open(page, '#/system/insights/heatmaps?map=' + MAP);
    await expect(root(page).locator('[data-economy-tab="heatmaps"]')).toHaveAttribute('aria-current', 'page');
    await expect(root(page).locator('[data-in-heat="' + MAP + '"] canvas')).toBeVisible();
    await expect(root(page).locator('[data-in-kpi="points"]')).toContainText('۲۴۲');
    await expect(root(page).locator('[data-in-kpi="uploads"]')).toContainText('۲');
    await root(page).locator('[data-in-hm-layer] [data-value="speed"]').click();
    await expect(root(page).locator('[data-in-heat="' + MAP + '"] canvas')).toBeVisible();
    await root(page).locator('[data-in-trace-btn]').first().click();
    await expect(root(page).locator('[data-in-trace]')).toBeVisible();
    await root(page).locator('[data-in-hm-edit]').click();
    const form = panel(page).locator('[data-in-map-form="' + MAP + '"]');
    await form.locator('[data-in-map="cellSize"]').fill('۸');
    await panel(page).locator('[data-in-map-save]').click();
    await expect(panel(page)).toHaveCount(0);
    expect((await base.get('/insights/heatmaps')).items.find((m: any) => m.key === MAP).cellSize).toBe(8);
  });

  test('settings save retention days, Ctrl K reaches the heatmaps', async ({ page }) => {
    await open(page, '#/system/insights/settings');
    const form = page.locator('[data-settings-form="insights"]');
    await expect(form).toBeVisible();
    const days = form.locator('input[dir="ltr"]').nth(1);
    await expect(days).toHaveValue('۱, ۷, ۳۰');
    await days.fill('۱, ۳, ۷');
    await form.locator('[data-settings-save]').click();
    await expect.poll(async () => (await base.get('/insights/settings')).settings.retention.days).toEqual([1, 3, 7]);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('هیت‌مپ');
    await expect(page.locator('.pl-pal-row').first().locator('.pl-pal-label')).toHaveText('هیت‌مپ');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/system\/insights\/heatmaps/);
  });

  test('English, dark, 390px', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/system/insights?platform=' + TAG);
    await expect(root(m).locator('h1')).toHaveText('Analytics');
    await expect(root(m).locator('[data-in-kpi="dau"]')).toContainText('2');
    await noHorizontalScroll(m);
    await open(m, '#/system/insights/heatmaps?map=' + MAP);
    await expect(root(m).locator('[data-in-heat="' + MAP + '"] canvas')).toBeVisible();
    await noHorizontalScroll(m);
    await open(m, '#/system/insights/sessions');
    await expect(root(m).locator('[data-economy-tab="insights-sessions"]')).toHaveAttribute('aria-current', 'page');
    await noHorizontalScroll(m);
  });
});
