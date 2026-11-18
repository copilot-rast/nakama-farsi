import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, rpc, uid, type Player } from '../../lib/api';
import { cup, rawCupSettings, ready, removeCups, report, restoreCupSettings } from '../../lib/cup';
import { noHorizontalScroll } from '../../lib/console';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const part = 'e2e' + k;
const F: Record<string, any> = {};
const players: Player[] = [];
let raw = '';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#cup-root');
const panel = (page: Page) => page.locator('[data-cup=panel] [role=dialog]');
const toasts = (page: Page) => page.locator('[data-cup=toasts] .eco-notice');
const fa = (n: number) => String(n).replace(/\d/g, (d) => '۰۱۲۳۴۵۶۷۸۹'[Number(d)]);

async function openPage(page: Page, sub: string, id = sub.split('?')[0]) {
  await page.goto(CONSOLE_URL + '/#/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-cup-page="${id}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

async function confirm(page: Page) {
  await page.locator('[data-cup=confirm] [data-cup-confirm=ok]').click();
}

async function detail(id: string) {
  return cup.get('/list/' + id);
}

test.beforeAll(async () => {
  raw = rawCupSettings();
  const s = (await cup.get('/settings')).settings;
  s.enabled = true;
  s.events.notify = false;
  s.events.inApp = false;
  s.events.publish = false;
  s.limits.activePerPlayer = 0;
  await cup.put('/settings', s);
  for (let i = 0; i < 7; i++) players.push(await devicePlayer('e2e-cup-' + k + '-' + i, part + 'p' + i));
});

test.afterAll(async () => {
  removeCups(part);
  for (const p of players) await deleteUser(p.id);
  if (raw) {
    await cup.raw('PUT', '/settings', JSON.parse(raw));
    restoreCupSettings(raw);
  }
});

test.describe('cup console', () => {
  test('sidebar group and Ctrl K; a cup is created with a best-of-3 final, a bad capacity is refused', async ({ page }) => {
    await openPage(page, 'cups/list');
    const g = page.locator('[data-cup-group="cups"]');
    await expect(g.locator('a[href="#/cups/list"]')).toBeAttached();
    await expect(g.locator('a[href="#/cups/templates"]')).toBeAttached();
    await expect(g.locator('a[href="#/cups/settings"]')).toBeAttached();
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('تورنمنت‌های تکراری');
    await expect(page.locator('.pl-pal-row').first()).toContainText('تورنمنت‌های تکراری');
    await page.keyboard.press('Escape');

    await root(page).locator('[data-cup-action="new-cup"]').click();
    const p = panel(page);
    await expect(p.locator('[data-cup-section="timing"]')).toBeVisible();
    await p.locator('input[name="name-fa"]').fill('جام ' + part);
    await p.locator('input[name="rules.capacity"]').fill('1');
    await p.locator('[data-cup-action="save-cup"]').click();
    await expect(p.locator('[data-cup-msg] .eco-notice')).toContainText('rules.capacity');
    await p.locator('input[name="rules.capacity"]').fill(fa(8));
    await p.locator('input[name="rules.minPlayers"]').fill('2');
    await p.locator('[data-cup-add="roundBestOf"]').click();
    await p.locator('select[name="roundBestOf.round"]').selectOption('final');
    await p.locator('select[name="roundBestOf.bestOf"]').selectOption('3');
    await p.locator('select[name="rules.timing.mode"]').selectOption('live');
    await expect(p.locator('input[name="rules.timing.checkInMinutes"]')).toBeVisible();
    await expect(p.locator('input[name="rules.timing.windowMinutes"]')).toBeHidden();
    await p.locator('select[name="rules.timing.mode"]').selectOption('window');
    await p.locator('input[name="rules.thirdPlace"]').check();
    await p.locator('[data-cup-action="save-cup"]').click();
    await expect(p).toHaveCount(0);
    await expect(page).toHaveURL(/#\/cups\/list\/[0-9a-f-]{36}$/);
    F.id = /#\/cups\/list\/([0-9a-f-]{36})/.exec(page.url())![1];
    await expect(root(page).locator('[data-cup-value="status"]')).toContainText('پیش‌نویس');
    await expect(root(page).locator('h1')).toHaveText('جام ' + part);
    const d = await detail(F.id);
    expect(d.cup.rules).toMatchObject({ capacity: 8, minPlayers: 2, bestOf: 1, thirdPlace: true, roundBestOf: [{ round: 'final', bestOf: 3 }], timing: { mode: 'window' } });
    await root(page).locator('[data-cup-back]').click();
    await expect(root(page).locator('[data-cup-page="cups/list"]')).toBeVisible();
    await root(page).locator('input[name="q"]').fill(part);
    await expect(root(page).locator(`tr[data-cup-row="${F.id}"]`)).toContainText('پیش‌نویس');
  });

  test('publish, players sign up and one is added by username, a swap before the start', async ({ page }) => {
    await openPage(page, 'cups/list/' + F.id, 'cups/view');
    await root(page).locator('[data-cup-action="publish"]').click();
    await expect(root(page).locator('[data-cup-value="status"]')).toContainText('ثبت‌نام باز');
    for (const p of players.slice(0, 4)) {
      const r = await rpc(p, 'cup_signup', { cupId: F.id });
      expect(r.status, JSON.stringify(r.data)).toBe(200);
    }
    await root(page).locator('[data-cup-action="add-player"]').click();
    await panel(page).locator('input[name="userId"]').fill(players[4].username);
    await panel(page).locator('[data-cup-action="save-player"]').click();
    await expect(panel(page)).toHaveCount(0);
    await expect(toasts(page).last()).toContainText('بازیکن اضافه شد');
    const box = root(page).locator('[data-cup-section="players"]');
    await expect(box.locator('tbody tr')).toHaveCount(5);
    await expect(root(page).locator('[data-cup-value="players"]')).toContainText(fa(5) + ' از ' + fa(8));

    await box.locator(`tr[data-cup-row="${players[0].id}"] [data-cup-player-action="swap"]`).click();
    await panel(page).locator('input[name="to"]').fill(players[5].id);
    await panel(page).locator('[data-cup-action="save-swap"]').click();
    await expect(panel(page)).toHaveCount(0);
    await expect(box.locator(`tr[data-cup-row="${players[5].id}"]`)).toBeVisible();
    await expect(box.locator(`tr[data-cup-row="${players[0].id}"]`)).toHaveCount(0);
    const d = await detail(F.id);
    expect(d.players.find((x: any) => x.userId === players[0].id).status).toBe('removed');
  });

  test('start now builds an 8-slot bracket with byes; force a result, disqualify, best-of-3 final, champion and places', async ({ page }) => {
    await openPage(page, 'cups/list/' + F.id, 'cups/view');
    await root(page).locator('[data-cup-action="start"]').click();
    await confirm(page);
    await expect(root(page).locator('[data-cup-value="status"]')).toContainText('در حال برگزاری');
    const br = root(page).locator('[data-cup-bracket]');
    await expect(br.locator('[data-cup-round]')).toHaveCount(3);
    await expect(br.locator('[data-cup-round="1"] .cu-round-head')).toContainText('یک‌چهارم نهایی');
    await expect(br.locator('[data-cup-round="3"] .cu-round-head')).toContainText('بهترین از ۳');
    await expect(br.locator('[data-cup-round="1"] [data-cup-status="bye"]')).toHaveCount(3);
    let d = await detail(F.id);
    expect(d.cup).toMatchObject({ size: 8, rounds: 3, players: 5 });
    const r1 = d.bracket.rounds[0].matches.filter((m: any) => m.status !== 'bye');
    expect(r1.length).toBe(1);
    const m1 = r1[0];
    await br.locator(`[data-cup-match="${m1.id}"]`).click();
    const mp = panel(page);
    await expect(mp.locator('[data-cup-section="force"]')).toBeVisible();
    await mp.locator(`[data-cup-force="${m1.a.userId}"]`).click();
    await confirm(page);
    await expect(mp).toHaveCount(0);
    await expect(toasts(page).last()).toContainText('نتیجه ثبت شد');
    d = await detail(F.id);
    expect(d.bracket.rounds[0].matches.find((m: any) => m.id === m1.id)).toMatchObject({ status: 'finished', winnerId: m1.a.userId, reason: 'admin' });

    const semis = d.bracket.rounds[1].matches;
    expect(semis.every((m: any) => m.a && m.b)).toBe(true);
    const [s1, s2] = semis;
    const dq = s1.b.userId;
    const box = root(page).locator('[data-cup-section="players"]');
    await box.locator(`tr[data-cup-row="${dq}"] [data-cup-player-action="disqualify"]`).click();
    await panel(page).locator('[data-cup-action="save-disqualify"]').click();
    await expect(panel(page)).toHaveCount(0);
    await expect(box.locator(`tr[data-cup-row="${dq}"]`)).toContainText('محروم');
    d = await detail(F.id);
    expect(d.bracket.rounds[1].matches[0]).toMatchObject({ status: 'finished', winnerId: s1.a.userId, reason: 'disqualified' });

    const byId = (id: string) => players.find((p) => p.id === id)!;
    await ready(byId(s2.a.userId), F.id);
    await ready(byId(s2.b.userId), F.id);
    await expect.poll(async () => (await detail(F.id)).bracket.rounds[1].matches[1].status).toBe('playing');
    d = await detail(F.id);
    const live = d.bracket.rounds[1].matches[1];
    expect(live.gameMatchId).toMatch(/^cup:/);
    expect((await report({ gameMatchId: live.gameMatchId, winnerId: s2.b.userId, present: [s2.a.userId, s2.b.userId] })).status).toBe(200);

    d = await detail(F.id);
    const fin = d.bracket.rounds[2].matches[0];
    const third = d.bracket.third;
    expect(fin.bestOf).toBe(3);
    expect([fin.a.userId, fin.b.userId].sort()).toEqual([s1.a.userId, s2.b.userId].sort());
    expect(third.b ? third.b.userId : third.a.userId).toBeTruthy();
    for (let game = 1; game <= 2; game++) {
      await ready(byId(fin.a.userId), F.id);
      await ready(byId(fin.b.userId), F.id);
      await expect.poll(async () => (await detail(F.id)).bracket.rounds[2].matches[0].status).toBe('playing');
      const cur = (await detail(F.id)).bracket.rounds[2].matches[0];
      const res = await report({ gameMatchId: cur.gameMatchId, winnerId: fin.a.userId, present: [fin.a.userId, fin.b.userId] });
      expect(res.status, JSON.stringify(res.data)).toBe(200);
    }
    d = await detail(F.id);
    expect(d.bracket.rounds[2].matches[0]).toMatchObject({ status: 'finished', winnerId: fin.a.userId });
    expect(d.bracket.rounds[2].matches[0].a.wins + d.bracket.rounds[2].matches[0].b.wins).toBe(2);
    const t3 = d.bracket.third;
    if (t3.status !== 'finished') {
      await page.reload();
      await root(page).locator(`[data-cup-match="${t3.id}"]`).click();
      await panel(page).locator(`[data-cup-force="${t3.a.userId}"]`).click();
      await confirm(page);
      await expect(panel(page)).toHaveCount(0);
    }
    await expect(root(page).locator('[data-cup-value="status"]')).toContainText('تمام‌شده', { timeout: 15_000 });
    await expect(root(page).locator('[data-cup-value="champion"]')).toContainText(byId(fin.a.userId).username);
    d = await detail(F.id);
    expect(d.cup.status).toBe('finished');
    expect(d.cup.championId).toBe(fin.a.userId);
    expect(d.players.find((x: any) => x.userId === fin.a.userId).place).toBe(1);
    expect(d.players.find((x: any) => x.userId === fin.b.userId).place).toBe(2);
    await expect(root(page).locator('[data-cup-action="cancel"]')).toHaveCount(0);
  });

  test('cancel with refund from the bracket page, then delete', async ({ page }) => {
    const now = Date.now();
    const out = await cup.post('/list', { name: { fa: 'لغو ' + part }, signupAt: new Date(now - 60_000).toISOString(), startsAt: new Date(now + 3_600_000).toISOString(), publish: true,
      rules: { ...(await cup.get('/refs')).defaults, capacity: 8, minPlayers: 2 } });
    F.cancel = out.cup.id;
    expect((await rpc(players[6], 'cup_signup', { cupId: F.cancel })).status).toBe(200);
    await openPage(page, 'cups/list/' + F.cancel, 'cups/view');
    await root(page).locator('[data-cup-action="cancel"]').click();
    await panel(page).locator('input[name="reason"]').fill('آزمایش');
    await panel(page).locator('[data-cup-action="save-cancel"]').click();
    await expect(panel(page)).toHaveCount(0);
    await expect(root(page).locator('[data-cup-value="status"]')).toContainText('لغوشده');
    await expect(root(page).locator('[data-cup-value="cancel-reason"]')).toHaveText('آزمایش');
    const d = await detail(F.cancel);
    expect(d.cup).toMatchObject({ status: 'cancelled', cancelReason: 'آزمایش' });
    expect((await rpc(players[6], 'cup_get', { cupId: F.cancel })).data.cup.status).toBe('cancelled');
    await root(page).locator('[data-cup-action="delete"]').click();
    await confirm(page);
    await expect(root(page).locator('[data-cup-page="cups/list"]')).toBeVisible();
    expect((await cup.raw('GET', '/list/' + F.cancel)).status).toBe(404);
  });

  test('recurring template: create, run now, delete', async ({ page }) => {
    await openPage(page, 'cups/templates');
    await root(page).locator('[data-cup-action="new-template"]').click();
    const p = panel(page);
    await p.locator('input[name="key"]').fill(part + '_weekly');
    await p.locator('input[name="name-fa"]').fill('هفتگی ' + part);
    await p.locator('input[name="schedule"]').fill('bad cron');
    await p.locator('[data-cup-action="save-template"]').click();
    await expect(p.locator('[data-cup-msg] .eco-notice')).toContainText('CRON');
    await p.locator('input[name="schedule"]').fill('0 18 * * 5');
    await p.locator('input[name="rules.capacity"]').fill('16');
    await p.locator('[data-cup-action="save-template"]').click();
    await expect(p).toHaveCount(0);
    const row = root(page).locator(`tr[data-cup-row="${part}_weekly"]`);
    await expect(row).toContainText('غیرفعال');
    await expect(row).toContainText('0 18 * * 5');
    await row.click();
    await panel(page).locator('[data-cup-action="run-template"]').click();
    await confirm(page);
    await expect(page).toHaveURL(/#\/cups\/list\/[0-9a-f-]{36}$/);
    const id = /#\/cups\/list\/([0-9a-f-]{36})/.exec(page.url())![1];
    const d = await detail(id);
    expect(d.cup).toMatchObject({ templateKey: part + '_weekly', rules: { capacity: 16 } });
    await expect(root(page).locator('h1')).toHaveText('هفتگی ' + part);
    await openPage(page, 'cups/templates');
    await root(page).locator(`tr[data-cup-row="${part}_weekly"]`).click();
    await panel(page).locator('[data-cup-action="delete-template"]').click();
    await confirm(page);
    await expect(root(page).locator(`tr[data-cup-row="${part}_weekly"]`)).toHaveCount(0);
  });

  test('settings save and keep the default rules', async ({ page }) => {
    await openPage(page, 'cups/settings');
    await page.locator('input[name="display.listLimit"]').fill('25');
    await page.locator('select[name="rules.bestOf"]').selectOption('3');
    await page.locator('[data-cup-save="cups/settings"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    const s = (await cup.get('/settings')).settings;
    expect(s.display.listLimit).toBe(25);
    expect(s.defaults.bestOf).toBe(3);
  });

  test('390px English dark: list, bracket, templates and settings without sideways scroll', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['list', 'Tournaments'], ['templates', 'Recurring tournaments'], ['settings', 'Tournament settings']]) {
      await m.goto(CONSOLE_URL + '/#/cups/' + sub);
      await expect(m.locator(`#cup-root [data-cup-page="cups/${sub}"] h1`)).toHaveText(title);
      await expect(m.getByText('Loading…')).toHaveCount(0);
      await noHorizontalScroll(m);
    }
    await m.goto(CONSOLE_URL + '/#/cups/list/' + F.id);
    await expect(m.locator('#cup-root [data-cup-value="status"]')).toContainText('Finished');
    await expect(m.locator('#cup-root [data-cup-bracket] [data-cup-round]')).toHaveCount(3);
    await noHorizontalScroll(m);
    await m.locator('#cup-root [data-cup-bracket] [data-cup-round="3"] [data-cup-match]').scrollIntoViewIfNeeded();
    await m.locator('#cup-root [data-cup-bracket] [data-cup-round="3"] [data-cup-match]').click();
    await expect(m.locator('[data-cup=panel] [data-cup-section="games"] [data-cup-game]')).toHaveCount(2);
    await noHorizontalScroll(m);
  });
});
