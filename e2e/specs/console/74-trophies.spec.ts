import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, uid, type Player } from '../../lib/api';
import { rawSettings, removeTrophyPart, removeTrophyPlayer, restoreSettings, tro } from '../../lib/clan';
import { noHorizontalScroll } from '../../lib/console';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const part = 'e2e_' + k;
const LG = part + '_gold';
const MIN = 1_800_000 + Math.floor(Math.random() * 90_000);
const F: Record<string, any> = {};
const users: string[] = [];
let raw = '';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#clan-root');
const panel = (page: Page) => page.locator('[data-clan=panel] [role=dialog]');
const toasts = (page: Page) => page.locator('[data-clan=toasts] .eco-notice');
const fa = (n: number) => n.toLocaleString('en-US').replace(/\d/g, (d) => '۰۱۲۳۴۵۶۷۸۹'[Number(d)]).replace(/,/g, '٬');

async function openPage(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-clan-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

test.beforeAll(async () => {
  raw = rawSettings('trophy_settings');
  await tro.setSettings((s) => { s.enabled = true; s.events.publish = false; s.events.notify = false; });
  const p: Player = await devicePlayer('e2e-tro-' + k);
  users.push(p.id);
  F.p = p;
});

test.afterAll(async () => {
  removeTrophyPart(part);
  for (const id of users) { removeTrophyPlayer(id); await deleteUser(id); }
  if (raw) {
    await tro.raw('PUT', '/settings', JSON.parse(raw));
    restoreSettings('trophy_settings', raw);
  }
});

test.describe('trophy console', () => {
  test('sidebar groups; a league is created, a bad key and a taken minimum are refused, then it is edited', async ({ page }) => {
    await openPage(page, 'trophies/leagues');
    await expect(page.locator('[data-clan-group="trophies"] a[href="#/trophies/players"]')).toBeAttached();
    await expect(page.locator('[data-clan-group="clans"] a[href="#/clans/rounds"]')).toBeAttached();
    await root(page).locator('[data-clan-action="new-league"]').click();
    let p = panel(page);
    await p.locator('input[name="key"]').fill('Bad Key!');
    await p.locator('input[name="name-fa"]').fill('طلایی ' + k);
    await p.locator('input[name="minTrophies"]').fill(fa(MIN).replace(/٬/g, ''));
    await p.locator('[data-clan-action="save-league"]').click();
    await expect(p.locator('.eco-notice')).toContainText('کلید یا نام درست نیست');
    await p.locator('input[name="key"]').fill(LG);
    await p.locator('input[name="floor"]').check();
    await p.locator('input[name="color"]').fill('#F5B301');
    await p.locator('[data-clan-action="save-league"]').click();
    await expect(p).toHaveCount(0);
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    const row = root(page).locator(`tr[data-clan-row="${LG}"]`);
    await expect(row).toContainText('از ' + fa(MIN) + ' جام');
    await expect(row).toContainText('پایین نمی‌افتد');
    F.league = (await tro.get('/leagues')).items.find((l: any) => l.key === LG);
    expect(F.league).toMatchObject({ minTrophies: MIN, floor: true, resetTo: null, isActive: true, color: '#F5B301', name: { fa: 'طلایی ' + k } });

    await root(page).locator('[data-clan-action="new-league"]').click();
    p = panel(page);
    await p.locator('input[name="key"]').fill(part + '_copy');
    await p.locator('input[name="name-fa"]').fill('کپی ' + k);
    await p.locator('input[name="minTrophies"]').fill(String(MIN));
    await p.locator('[data-clan-action="save-league"]').click();
    await expect(p.locator('.eco-notice')).toContainText('لیگی با همین کلید یا همین حداقل جام هست');
    await p.locator('input[name="key"]').fill(LG);
    await p.locator('input[name="minTrophies"]').fill(String(MIN + 1));
    await p.locator('[data-clan-action="save-league"]').click();
    await expect(p.locator('.eco-notice')).toContainText('· key');
    await p.getByRole('button', { name: 'انصراف' }).click();

    await row.click();
    p = panel(page);
    await expect(p.locator('input[name="key"]')).toHaveValue(LG);
    await p.locator('input[name="name-en"]').fill('Gold ' + k);
    await p.locator('input[name="resetTo"]').fill('۱۵۰۰');
    await p.locator('[data-clan-action="save-league"]').click();
    await expect(p).toHaveCount(0);
    await expect(row).toContainText('آخر فصل: ۱٬۵۰۰');
    expect((await tro.get('/leagues')).items.find((l: any) => l.key === LG)).toMatchObject({ resetTo: 1500, name: { fa: 'طلایی ' + k, en: 'Gold ' + k } });
  });

  test('a player found by ID gets trophies set and added; the log, the league and the list follow', async ({ page }) => {
    await openPage(page, 'trophies/players');
    await root(page).locator('input[name="q"]').fill(F.p.id);
    await root(page).locator('[data-clan-action="open-player"]').click();
    const p = panel(page);
    await expect(p.locator('[data-clan-value="trophies"]')).toHaveText('۰');
    await p.locator('input[name="trophies"]').fill('۲۰۰۰۰۰۰');
    await p.locator('select[name="mode"]').selectOption('set');
    await p.locator('[data-clan-action="set-trophies"]').click();
    await expect(p.locator('[data-clan-value="trophies"]')).toHaveText(fa(2_000_000));
    await expect(p.locator('[data-clan-dl="trophy-player"]')).toContainText('طلایی ' + k);
    const log = p.locator('[data-clan-section="log"] tbody tr');
    await expect(log).toHaveCount(1);
    await expect(log.first()).toContainText('ادمین');
    await expect(log.first()).toContainText('+' + fa(2_000_000));
    await p.locator('input[name="trophies"]').fill('-50');
    await p.locator('select[name="mode"]').selectOption('add');
    await p.locator('[data-clan-action="set-trophies"]').click();
    await expect(p.locator('[data-clan-value="trophies"]')).toHaveText(fa(1_999_950));
    await expect(log).toHaveCount(2);
    await expect(log.first()).toContainText('−۵۰');
    const d = await tro.get('/players/' + F.p.id);
    expect(d).toMatchObject({ trophies: 1_999_950, best: 2_000_000, league: LG });
    expect(d.log[0]).toMatchObject({ reason: 'admin', delta: -50, before: 2_000_000, after: 1_999_950 });
    await page.keyboard.press('Escape');
    await expect(root(page).locator(`tr[data-clan-row="${F.p.id}"]`)).toContainText(fa(1_999_950));
    await root(page).locator('select[name="league"]').selectOption(LG);
    await expect(root(page).locator('[data-clan-list="trophy-players"] tbody tr')).toHaveCount(1);
  });

  test('ending the season now cuts trophies by the rule and shows in the list and the player log', async ({ page }) => {
    await tro.setSettings((s) => { s.season.mode = 'cut'; s.season.above = 1_000_000; s.season.keepPercent = 50; });
    await openPage(page, 'trophies/seasons');
    const before = (await tro.get('/seasons')).items.length;
    await expect(root(page).locator('[data-clan-dl="season"]')).toContainText('کم کردن جام بالای یک عدد');
    await root(page).locator('[data-clan-action="reset-season"]').click();
    await expect(page.locator('[data-clan=confirm]')).toContainText('کم کردن جام بالای یک عدد');
    await page.locator('[data-clan-confirm="ok"]').click();
    await expect(toasts(page).last()).toContainText('فصل تمام شد');
    const after = await tro.get('/seasons');
    expect(after.items.length).toBeGreaterThan(before);
    expect(after.items.find((s: any) => s.endedAt && s.actor)).toBeTruthy();
    await expect(root(page).locator('[data-clan-list="seasons"] tbody tr').first()).toContainText('فصل جاری');
    const d = await tro.get('/players/' + F.p.id);
    expect(d.trophies).toBe(1_000_000 + Math.floor(999_950 * 50 / 100));
    expect(d.log[0]).toMatchObject({ reason: 'season' });
    await openPage(page, 'trophies/players');
    await root(page).locator('input[name="q"]').fill(F.p.id);
    await root(page).locator(`tr[data-clan-row="${F.p.id}"]`).click();
    await expect(panel(page).locator('[data-clan-section="log"] tbody tr').first()).toContainText('پایان فصل');
  });

  test('settings: a bad schedule names its field, a good save sticks and the original comes back', async ({ page }) => {
    const orig = await tro.settings();
    await openPage(page, 'trophies/settings');
    const box = root(page);
    await box.locator('input[name="amounts.win"]').fill('۴۵');
    await box.locator('input[name="season.schedule"]').fill('nope');
    await box.locator('[data-clan-save="trophies/settings"]').click();
    await expect(box.locator('[data-clan-msg]')).toContainText('زمان‌بندی (cron) درست نیست');
    await expect(box.locator('[data-clan-msg]')).toContainText('season.schedule');
    await box.locator('input[name="season.schedule"]').fill('0 0 1 * *');
    await box.locator('input[name="match.modes"]').fill('ranked، cup');
    await box.locator('[data-clan-save="trophies/settings"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    const s = await tro.settings();
    expect(s.amounts.win).toBe(45);
    expect(s.match.modes).toEqual(['ranked', 'cup']);
    expect(s.season.schedule).toBe('0 0 1 * *');
    await tro.put('/settings', orig);
    await page.reload();
    await expect(root(page).locator('input[name="amounts.win"]')).toHaveValue(String(orig.amounts.win).replace(/\d/g, (x) => '۰۱۲۳۴۵۶۷۸۹'[Number(x)]));
  });

  test('a league is deleted after a confirm', async ({ page }) => {
    await openPage(page, 'trophies/leagues');
    await root(page).locator(`[data-clan-delete="${LG}"]`).click();
    await page.locator('[data-clan-confirm="ok"]').click();
    await expect(root(page).locator(`tr[data-clan-row="${LG}"]`)).toHaveCount(0);
    expect((await tro.get('/leagues')).items.some((l: any) => l.key === LG)).toBe(false);
  });

  test('Ctrl K finds the trophy pages; 390px English dark without sideways scroll', async ({ page, spawn }) => {
    await openPage(page, 'trophies/leagues');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('جام بازیکنان');
    await expect(page.locator('.pl-pal-row').first()).toContainText('جام بازیکنان');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/trophies\/players/);
    await expect(root(page).locator('[data-clan-page="trophies/players"]')).toBeVisible();

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['leagues', 'Leagues'], ['players', 'Player trophies'], ['seasons', 'Seasons'], ['settings', 'Trophy settings']]) {
      await m.goto(CONSOLE_URL + '/#/trophies/' + sub);
      await expect(m.locator(`#clan-root [data-clan-page="trophies/${sub}"] h1`)).toHaveText(title);
      await expect(m.getByText('Loading…')).toHaveCount(0);
      await noHorizontalScroll(m);
    }
    await m.goto(CONSOLE_URL + '/#/trophies/players');
    await m.locator('#clan-root input[name="q"]').fill(F.p.id);
    await m.locator(`#clan-root tr[data-clan-row="${F.p.id}"]`).click();
    await expect(m.locator('[data-clan=panel] [data-clan-section="log"]')).toContainText('Season end');
    await expect(m.locator('[data-clan=panel] [data-clan-value="trophies"]')).toHaveText(/^[0-9,]+$/);
    await noHorizontalScroll(m);
  });
});
