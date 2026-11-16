import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { lbx, removeBoards } from '../../lib/leaderboardx';
import { noHorizontalScroll } from '../../lib/console';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const part = 'e2e_' + k;
const KEY = part + '_season';
const F: Record<string, any> = {};
const users: string[] = [];
let rawSettings = '';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#lbx-root');
const panel = (page: Page) => page.locator('[data-lbx=panel] [role=dialog]');

async function openLb(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/leaderboardx/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-lbx-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM leaderboardx_settings WHERE id=1`).trim();
  await lbx.setSettings((s) => { s.events.economyEvent = ''; s.events.analyticsEvent = ''; s.rewards.payMode = 'auto'; });
  F.cur = 'e2e' + k + 'c';
  await eco.post('/currencies', { key: F.cur, name: { fa: 'سکهٔ فصل ' + k, en: 'Season coin ' + k }, type: 'soft', defaultValue: 0, iconUrl: '', tagIds: [], metadata: {}, sort: 99 });
  F.prize = await eco.post('/prizes', { key: 'e2e_' + k + '_cup', name: { fa: 'جام ' + k }, message: {}, tagIds: [], isRepeatable: true, metadata: {}, items: [{ type: 'currency', key: F.cur, action: 'add', value: 1 }] });
  for (const tag of ['a', 'b']) {
    const p: Player = await devicePlayer(`e2e-lbc-${tag}-${k}`);
    users.push(p.id);
    F[tag] = p;
  }
});

test.afterAll(async () => {
  await removeBoards(part);
  for (const id of users) await deleteUser(id);
  if (rawSettings) sql(`UPDATE leaderboardx_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
  await lbx.raw('GET', '/settings');
  if (F.prize) { sql(`DELETE FROM economy_user_prize WHERE prize_id='${F.prize.id}'`); await eco.del('/prizes/' + F.prize.id); }
  const c = (await eco.get('/currencies?q=' + F.cur)).items?.find((x: any) => x.key === F.cur);
  if (c) await eco.del('/currencies/' + c.id);
});

test.describe('leaderboardx console', () => {
  test('sidebar group; a board with tiers and reward rows from the editor; a wrong tier order shows on the form', async ({ page }) => {
    await openLb(page, 'boards');
    await expect(page.locator('[data-lbx-group] a[href="#/leaderboardx/winners"]')).toBeAttached();
    await root(page).locator('[data-lbx-action="new-board"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill(KEY);
    await p.locator('input[name="name-fa"]').fill('فصل گشت ' + k);
    await p.locator('select[name="period"]').selectOption('custom');
    await p.locator('input[name="resetSchedule"]').fill('0 0 * * 5');
    await p.locator('select[name="source.type"]').selectOption('client');
    await expect(p.locator('input[name="limits.clientSubmit"]')).toBeChecked();
    await p.locator('input[name="limits.maxPerWrite"]').fill('۵۰۰');
    await p.locator('select[name="tierBasis"]').selectOption('score');
    await p.locator('[data-lbx-action="add-tier"]').click();
    await p.locator('[data-lbx-action="add-tier"]').click();
    const tiers = p.locator('[data-lbx-tier]');
    await tiers.nth(0).locator('input[name="tier-key"]').fill('gold');
    await tiers.nth(0).locator('input[name="tier-threshold"]').fill('10');
    await tiers.nth(0).locator('input[name="tier-name-fa"]').fill('طلایی');
    await tiers.nth(1).locator('input[name="tier-key"]').fill('silver');
    await tiers.nth(1).locator('input[name="tier-threshold"]').fill('100');
    await expect(tiers.nth(1).locator('[data-lbx-tier-sentence]')).toHaveText('امتیاز ۱۰۰ یا بهتر');
    await p.locator('[data-lbx-action="add-reward"]').click();
    const rw = p.locator('[data-lbx-reward]').first();
    await rw.locator('input[name="reward-to"]').fill('۲');
    await expect(rw.locator('[data-lbx-reward-sentence]')).toHaveText('رتبهٔ ۱ تا ۲');
    await rw.locator('[data-lbx-action="add-currency"]').click();
    await rw.locator('select[name="reward-currency"]').selectOption(F.cur);
    await rw.locator('input[name="reward-amount"]').fill('30');
    await p.locator('[data-lbx-action="add-reward"]').click();
    const rw2 = p.locator('[data-lbx-reward]').nth(1);
    await rw2.locator('select[name="reward-tier"]').focus();
    await rw2.locator('select[name="reward-tier"]').selectOption('silver');
    await rw2.locator('input[name="reward-to"]').fill('');
    await rw2.locator('select[name="reward-prize"]').selectOption(F.prize.key);
    await expect(rw2.locator('[data-lbx-reward-sentence]')).toContainText('همهٔ سطح');
    await p.locator('[data-lbx-action="save-board"]').click();
    await expect(p.locator('[data-lbx-msg]')).toContainText('سطح‌ها باید از بهترین به پایین مرتب باشند');
    await tiers.nth(0).locator('input[name="tier-threshold"]').fill('100');
    await tiers.nth(1).locator('input[name="tier-threshold"]').fill('10');
    await p.locator('[data-lbx-action="save-board"]').click();
    await expect(p).toHaveCount(0);
    const row = root(page).locator(`tr[data-lbx-row="${KEY}"]`);
    await expect(row).toContainText('فصل گشت');
    await expect(row.locator(`[data-lbx-rewards="${KEY}"]`)).toContainText('رتبهٔ ۱ تا ۲');
    const b = (await lbx.get('/boards')).items.find((x: any) => x.board.key === KEY).board;
    expect(b).toMatchObject({ period: 'custom', resetSchedule: '0 0 * * 5', tierBasis: 'score', limits: { clientSubmit: true, maxPerWrite: 500 },
      tiers: [{ key: 'gold', name: { fa: 'طلایی' }, threshold: 100 }, { key: 'silver', threshold: 10 }] });
    expect(b.rewards).toEqual([{ fromRank: 1, toRank: 2, currencies: { [F.cur]: 30 } }, { tier: 'silver', fromRank: 1, toRank: 0, prizeKey: F.prize.key }]);
    expect(sql(`SELECT reset_schedule FROM leaderboard WHERE id='${KEY}'`).trim()).toBe('0 0 * * 5');
    F.board = b;
  });

  test('live standings: a score changed by hand, tiers shown; closing the period pays and lists the winners', async ({ page }) => {
    for (const [p, s] of [[F.a, 150], [F.b, 40]] as [Player, number][]) {
      const r = await rpc(p, 'leaderboardx_submit', { key: KEY, score: s, requestId: 'c' + s });
      expect(r.status, JSON.stringify(r.data)).toBe(200);
    }
    await openLb(page, 'standings?board=' + KEY);
    await expect(root(page).locator('[data-lbx-total]')).toHaveText('۲ بازیکن');
    await expect(root(page).locator(`[data-lbx-rank="${F.a.id}"]`)).toHaveText('۱');
    await expect(root(page).locator(`tr[data-lbx-row="${F.b.id}"]`)).toContainText('silver');
    await root(page).locator(`[data-lbx-edit="${F.b.id}"]`).click();
    const p = panel(page);
    await p.locator('input[name="score"]').fill('۲۰۰');
    await p.locator('[data-lbx-action="save-score"]').click();
    await expect(p).toHaveCount(0);
    await expect(root(page).locator(`[data-lbx-rank="${F.b.id}"]`)).toHaveText('۱');
    await expect(root(page).locator(`tr[data-lbx-row="${F.b.id}"]`)).toContainText('طلایی');
    await root(page).locator('[data-lbx-action="close-period"]').click();
    await page.locator('[data-lbx-confirm="ok"]').click();
    await expect(page.getByText('دوره بسته شد')).toBeVisible();
    await expect(root(page).locator('[data-lbx-total]')).toHaveText('۰ بازیکن');
    const bal = (id: string) => Number(sql(`SELECT COALESCE((wallet->>'${F.cur}')::bigint,0) FROM users WHERE id='${id}'`).trim());
    expect(bal(F.a.id)).toBe(30);
    expect(bal(F.b.id)).toBe(30);
    expect(sql(`SELECT count(*) FROM economy_user_prize WHERE source_type='leaderboard' AND user_id='${F.b.id}'`).trim()).toBe('0');

    await openLb(page, 'winners?board=' + KEY);
    const pid = sql(`SELECT id FROM leaderboardx_period WHERE board_key='${KEY}'`).trim();
    await expect(root(page).locator(`[data-lbx-period-players="${pid}"]`)).toHaveText('۲');
    await expect(root(page).locator(`[data-lbx-period-counts="${pid}"]`)).toContainText('پرداخت شد ۲');
    await root(page).locator(`tr[data-lbx-row="${pid}"]`).click();
    const pp = panel(page);
    await expect(pp.locator(`[data-lbx-result="${F.b.id}"]`)).toContainText('طلایی');
    await expect(pp.locator(`[data-lbx-result="${F.b.id}"] [data-lbx-status="paid"]`)).toBeVisible();
    F.pid = pid;
  });

  test('changing the reset period asks to close the running period first', async ({ page }) => {
    const r = await rpc(F.a, 'leaderboardx_submit', { key: KEY, score: 12, requestId: 'after' });
    expect(r.status).toBe(200);
    await openLb(page, 'boards');
    await root(page).locator(`tr[data-lbx-row="${KEY}"]`).click();
    const p = panel(page);
    await p.locator('select[name="period"]').selectOption('daily');
    await p.locator('[data-lbx-action="save-board"]').click();
    await expect(p.locator('[data-lbx-restart]')).toContainText('۱ بازیکن');
    await p.locator('[data-lbx-action="restart"]').click();
    await expect(p).toHaveCount(0);
    expect(sql(`SELECT reason || ':' || rewarded FROM leaderboardx_period WHERE board_key='${KEY}' ORDER BY created_at DESC LIMIT 1`).trim()).toBe('restart:0');
    expect(sql(`SELECT reset_schedule FROM leaderboard WHERE id='${KEY}'`).trim()).toBe('0 0 * * *');
  });

  test('report and player card', async ({ page }) => {
    await openLb(page, 'report?days=7');
    await expect(root(page).locator(`[data-lbx-report-writes="${KEY}"]`)).not.toHaveText('۰');
    await expect(root(page).locator('[data-lbx-kpi="paid"] .text-xl')).not.toHaveText('۰');
    await page.goto(CONSOLE_URL + '/#/players/' + F.b.id);
    const card = page.locator('[data-lbx-player]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card.locator(`[data-lbx-player-result="${F.pid}"]`)).toContainText('رتبهٔ ۱');
    await expect(card.locator(`[data-lbx-player-result="${F.pid}"] [data-lbx-status="paid"]`)).toBeVisible();
  });

  test('settings are saved with checks; pages fit 390px in English dark mode', async ({ page, spawn }) => {
    await openLb(page, 'settings');
    await root(page).locator('select[name="rewards.payMode"]').selectOption('claim');
    await root(page).locator('input[name="rewards.claimDays"]').fill('۱۴');
    await root(page).locator('[data-lbx-save="settings"]').click();
    await expect(page.getByText('ذخیره شد')).toBeVisible();
    const s = await lbx.settings();
    expect(s.rewards).toMatchObject({ payMode: 'claim', claimDays: 14 });
    await root(page).locator('input[name="display.maxLimit"]').fill('5');
    await root(page).locator('[data-lbx-save="settings"]').click();
    await expect(root(page).locator('[data-lbx-msg]')).toContainText('عدد خارج از بازهٔ مجاز است');

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['boards', 'Boards'], ['standings', 'Live standings'], ['winners', 'Winners by period'], ['report', 'Leaderboard report'], ['settings', 'Leaderboard settings']]) {
      await m.goto(CONSOLE_URL + '/#/leaderboardx/' + sub);
      await expect(m.locator(`#lbx-root [data-lbx-page="${sub}"] h1`)).toHaveText(title);
      await noHorizontalScroll(m);
    }
  });
});
