import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, uid, type Player } from '../../lib/api';
import { cln, createClan, joinClan, rawSettings, removeClan, removeWarPart, restoreSettings } from '../../lib/clan';
import { lbx, removeBoards } from '../../lib/leaderboardx';
import { noHorizontalScroll } from '../../lib/console';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const short = k.slice(-7);
const part = 'e2e_' + k;
const WP = part + '_pts';
const WD = part + '_duel';
const BOARD = part + '_clanboard';
const F: Record<string, any> = {};
const users: string[] = [];
let raw = '';
let levels: any[] = [];

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#clan-root');
const panel = (page: Page) => page.locator('[data-clan=panel] [role=dialog]');
const toasts = (page: Page) => page.locator('[data-clan=toasts] .eco-notice');

async function openPage(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-clan-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

async function confirm(page: Page) {
  await page.locator('[data-clan-confirm="ok"]').click();
}

test.beforeAll(async () => {
  raw = rawSettings('clan_settings');
  levels = (await cln.get('/levels')).items.map((l: any) => ({ level: l.level, xp: l.xp, name: l.name, iconUrl: l.iconUrl, memberCap: l.memberCap, prizeKey: l.prizeKey, perks: l.perks }));
  await cln.setSettings((s) => {
    s.enabled = true; s.create.enabled = true; s.create.cost = {}; s.create.minTrophies = 0; s.events.publish = false; s.events.notify = false;
    s.cooldowns.joinAfterLeaveMinutes = 0; s.badge = { emblems: [], colors: [], frames: [] };
  });
  for (const tag of ['a', 'b', 'c']) {
    const p: Player = await devicePlayer(`e2e-cln-${tag}-${k}`);
    users.push(p.id);
    F[tag] = p;
  }
  F.one = (await createClan(F.a, 'e2e' + short + ' Lions')).clanId;
  F.two = (await createClan(F.c, 'e2e' + short + ' Tigers')).clanId;
  await joinClan(F.b, F.one);
});

test.afterAll(async () => {
  removeWarPart(part);
  await removeBoards(part);
  await removeClan(F.one);
  await removeClan(F.two);
  for (const id of users) await deleteUser(id);
  if (levels.length) await cln.raw('PUT', '/levels', { items: levels });
  if (raw) {
    await cln.raw('PUT', '/settings', JSON.parse(raw));
    restoreSettings('clan_settings', raw);
  }
});

test.describe('clan console', () => {
  test('clan list search and sort; the panel edits the profile and adds XP', async ({ page }) => {
    await openPage(page, 'clans/list');
    await expect(page.locator('[data-clan-group="clans"] a[href="#/clans/levels"]')).toBeAttached();
    await root(page).locator('input[name="q"]').fill('e2e' + short);
    await expect(root(page).locator('[data-clan-list="clans"] tbody tr')).toHaveCount(2);
    await root(page).locator('select[name="sort"]').selectOption('members');
    await expect(root(page).locator('[data-clan-list="clans"] tbody tr').first()).toHaveAttribute('data-clan-row', F.one);
    const row = root(page).locator(`tr[data-clan-row="${F.one}"]`);
    await expect(row).toContainText('۲ از');
    await expect(row).toContainText('آزاد');
    await row.click();
    const p = panel(page);
    await expect(p.locator('[data-clan-member]')).toHaveCount(2);
    await p.locator('[data-clan-form="profile"] textarea[name="description"]').fill('کلن آزمون ' + short);
    await p.locator('[data-clan-form="profile"] select[name="joinType"]').selectOption('closed');
    await p.locator('[data-clan-form="profile"] input[name="minTrophies"]').fill('۱۰');
    await p.locator('[data-clan-action="save-profile"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    await p.locator('input[name="xp-amount"]').fill('۵۰');
    await p.locator('[data-clan-action="add-xp"]').click();
    await expect(toasts(page).last()).toContainText('۵۰ امتیاز اضافه شد');
    await expect(p.locator('[data-clan-section="xp"] tbody tr').first()).toContainText('ادمین');
    await expect(p.locator('[data-clan-section="xp"] tbody tr').first()).toContainText('+۵۰');
    const d = await cln.get('/list/' + F.one);
    expect(d.clan).toMatchObject({ description: 'کلن آزمون ' + short, joinType: 'closed', minTrophies: 10, xp: 50 });
    await page.keyboard.press('Escape');
    await expect(row).toContainText('با درخواست');
  });

  test('member roles: promote to elder, the leader cannot be demoted, kick after a confirm', async ({ page }) => {
    await openPage(page, 'clans/list');
    await root(page).locator('input[name="q"]').fill('e2e' + short);
    await root(page).locator(`tr[data-clan-row="${F.one}"]`).click();
    const p = panel(page);
    const b = p.locator(`[data-clan-member="${F.b.id}"]`);
    await b.locator('select[name="role"]').selectOption('elder');
    await b.locator('[data-clan-action="role"]').click();
    await expect(p.locator(`[data-clan-member="${F.b.id}"] .eco-tag`)).toHaveText('ریش‌سفید');
    expect((await cln.get('/list/' + F.one)).members.find((m: any) => m.userId === F.b.id).role).toBe('elder');
    const a = p.locator(`[data-clan-member="${F.a.id}"]`);
    await a.locator('select[name="role"]').selectOption('member');
    await a.locator('[data-clan-action="role"]').click();
    await expect(toasts(page).last()).toContainText('رهبر را نمی‌شود پایین آورد');
    await p.locator(`[data-clan-member="${F.b.id}"] [data-clan-action="kick"]`).click();
    await confirm(page);
    await expect(p.locator('[data-clan-member]')).toHaveCount(1);
    await expect(p.locator('[data-clan-value="members"]')).toContainText('۱ از');
    const d = await cln.get('/list/' + F.one);
    expect(d.members.map((m: any) => m.userId)).toEqual([F.a.id]);
  });

  test('XP road: a level with too little XP is refused, a good one is saved, the original comes back', async ({ page }) => {
    await openPage(page, 'clans/levels');
    const rows = root(page).locator('[data-clan-level]');
    await expect(rows).toHaveCount(levels.length);
    await root(page).locator('[data-clan-action="add-level"]').click();
    await expect(rows).toHaveCount(levels.length + 1);
    const last = rows.last();
    await last.locator('input[name="xp"]').fill('0');
    await last.locator('input[name="name-fa"]').fill('سطح آزمون');
    await last.locator('input[name="memberCap"]').fill('۴۰');
    await root(page).locator('[data-clan-save="levels"]').click();
    await expect(root(page).locator('[data-clan-msg]')).toContainText('سطح‌ها تکراری نباشند');
    await last.locator('input[name="xp"]').fill('۹۹۹۹۹۹۹');
    await last.locator('textarea[name="perks"]').fill('{"chatSlots": 2}');
    await root(page).locator('[data-clan-save="levels"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    const items = (await cln.get('/levels')).items;
    expect(items[items.length - 1]).toMatchObject({ level: levels[levels.length - 1].level + 1, xp: 9_999_999, memberCap: 40, name: { fa: 'سطح آزمون' }, perks: { chatSlots: 2 } });
    await expect(root(page).locator('[data-clan-level]').last()).toContainText('۰ کلن در این سطح');
    await root(page).locator('[data-clan-level]').last().locator('[data-clan-action="remove-level"]').click();
    await root(page).locator('[data-clan-save="levels"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    expect((await cln.get('/levels')).items.length).toBe(levels.length);
  });

  test('war templates: a points war with a rule and a duel war with its fields', async ({ page }) => {
    await openPage(page, 'clans/wars');
    await root(page).locator('[data-clan-action="new-war"]').click();
    let p = panel(page);
    await p.locator('input[name="key"]').fill(WP);
    await p.locator('input[name="name-fa"]').fill('جنگ امتیازی ' + short);
    await expect(p.locator('[data-clan-rules="duel"]')).toBeHidden();
    await p.locator('[data-clan-action="save-war"]').click();
    await expect(p.locator('.eco-notice')).toContainText('جنگ امتیازی دست‌کم یک قانون امتیاز لازم دارد');
    await p.locator('[data-clan-action="add-point"]').click();
    const rule = p.locator('[data-clan-point]').first();
    await rule.locator('input[name="point-event"]').fill('e2e_kill');
    await rule.locator('input[name="point-field"]').fill('won');
    await rule.locator('input[name="point-equals"]').fill('true');
    await rule.locator('input[name="point-points"]').fill('۳');
    await p.locator('input[name="eligibility.minMembers"]').fill('1');
    await p.locator('[data-clan-action="save-war"]').click();
    await expect(p).toHaveCount(0);
    await expect(root(page).locator(`tr[data-clan-row="${WP}"]`)).toContainText('امتیازی');
    await expect(root(page).locator(`tr[data-clan-row="${WP}"]`)).toContainText('دستی');
    let c = (await cln.get('/wars/configs')).items.find((x: any) => x.key === WP);
    expect(c).toMatchObject({ mode: 'points', schedule: '', rules: { points: [{ type: 'economy_event', event: 'e2e_kill', field: 'won', equals: 'true', points: 3 }], eligibility: { minMembers: 1 } } });

    await root(page).locator('[data-clan-action="new-war"]').click();
    p = panel(page);
    await p.locator('input[name="key"]').fill(WD);
    await p.locator('input[name="name-fa"]').fill('جنگ دوئل ' + short);
    await p.locator('select[name="mode"]').selectOption('duel');
    await expect(p.locator('[data-clan-rules="duel"]')).toBeVisible();
    await expect(p.locator('[data-clan-rules="points"]')).toBeHidden();
    await expect(p.locator('input[name="duel.wonField"]')).toHaveValue('won');
    await p.locator('input[name="duel.event"]').fill('match_end');
    await p.locator('input[name="duel.winPoints"]').fill('۲');
    await p.locator('input[name="schedule"]').fill('* *');
    await p.locator('[data-clan-action="save-war"]').click();
    await expect(p.locator('.eco-notice')).toContainText('زمان‌بندی (cron) درست نیست');
    await p.locator('input[name="schedule"]').fill('');
    await p.locator('select[name="pairing"]').selectOption('trophies');
    await p.locator('[data-clan-action="save-war"]').click();
    await expect(p).toHaveCount(0);
    await expect(root(page).locator(`tr[data-clan-row="${WD}"]`)).toContainText('دوئل');
    c = (await cln.get('/wars/configs')).items.find((x: any) => x.key === WD);
    expect(c).toMatchObject({ mode: 'duel', pairing: 'trophies', rules: { duel: { type: 'economy_event', event: 'match_end', wonField: 'won', opponentField: 'opponentId', winPoints: 2 } } });

    await root(page).locator(`tr[data-clan-row="${WD}"]`).click();
    await expect(panel(page).locator('select[name="mode"]')).toHaveValue('duel');
    await expect(panel(page).locator('input[name="duel.winPoints"]')).toHaveValue('۲');
    await page.keyboard.press('Escape');
  });

  test('rounds: start now and settle; a later round is started then cancelled', async ({ page }) => {
    await openPage(page, 'clans/rounds');
    await root(page).locator('[data-clan-action="new-round"]').click();
    let p = panel(page);
    await p.locator('select[name="configKey"]').selectOption(WP);
    await expect(p.locator('input[name="startNow"]')).toBeChecked();
    await p.locator('[data-clan-action="create-round"]').click();
    p = page.locator('[data-clan=panel] [data-clan-panel="round"]');
    await expect(p.locator('[data-clan-value="round-status"]')).toContainText('در جریان');
    await expect(p.locator('[data-clan-section="wars"]')).toContainText('e2e' + short);
    await p.locator('[data-clan-action="settle-round"]').click();
    await confirm(page);
    await expect(p.locator('[data-clan-value="round-status"]')).toContainText('تمام‌شده');
    await expect(p.locator('[data-clan-action="cancel-round"]')).toHaveCount(0);
    let list = (await cln.get('/wars/rounds?configKey=' + WP)).items;
    expect(list[0].status).toBe('settled');
    const detail = await cln.get('/wars/rounds/' + list[0].id);
    expect(detail.items.some((w: any) => [w.clanA, w.clanB].includes(F.one))).toBe(true);
    await page.keyboard.press('Escape');

    await root(page).locator('[data-clan-action="new-round"]').click();
    p = panel(page);
    await p.locator('select[name="configKey"]').selectOption(WD);
    await p.locator('input[name="startNow"]').uncheck();
    await p.locator('[data-clan-action="create-round"]').click();
    await expect(p.locator('.eco-notice')).toContainText('زمان شروع را انتخاب کنید');
    const later = new Date(Date.now() + 3 * 86400_000);
    await p.locator('input[name="startsAt"]').fill(new Date(later.getTime() - later.getTimezoneOffset() * 60000).toISOString().slice(0, 16));
    await p.locator('[data-clan-action="create-round"]').click();
    p = page.locator('[data-clan=panel] [data-clan-panel="round"]');
    await expect(p.locator('[data-clan-value="round-status"]')).toContainText('ثبت‌نام');
    await expect(p.locator('[data-clan-section="wars"]')).toContainText('جنگ‌ها با شروع دور ساخته می‌شوند');
    await p.locator('[data-clan-action="start-round"]').click();
    await expect(p.locator('[data-clan-value="round-status"]')).toContainText('در جریان');
    await p.locator('[data-clan-action="cancel-round"]').click();
    await confirm(page);
    await expect(p.locator('[data-clan-value="round-status"]')).toContainText('لغوشده');
    list = (await cln.get('/wars/rounds?configKey=' + WD)).items;
    expect(list[0]).toMatchObject({ status: 'cancelled', mode: 'duel' });
    await page.keyboard.press('Escape');
    await root(page).locator('select[name="status"]').selectOption('cancelled');
    await expect(root(page).locator(`tr[data-clan-row="${list[0].id}"]`)).toContainText('لغوشده');
  });

  test('clan settings: role checkboxes, an XP source and cooldowns; a bad value names its field', async ({ page }) => {
    const orig = await cln.settings();
    await openPage(page, 'clans/settings');
    const box = root(page);
    const kickElder = box.locator('input[name="roles.kick.elder"]');
    if (await kickElder.isChecked()) await kickElder.uncheck(); else await kickElder.check();
    const before = await box.locator('[data-clan-xp-source]').count();
    await box.locator('[data-clan-action="add-xp-source"]').click();
    const src = box.locator('[data-clan-xp-source]').last();
    await src.locator('select[name="xp-type"]').selectOption('economy_event');
    await src.locator('input[name="xp-event"]').fill('e2e_win');
    await src.locator('input[name="xp-amount"]').fill('۷');
    await box.locator('[data-clan-action="add-cost"]').click();
    await box.locator('input[name="cooldowns.inviteDays"]').fill('0');
    await box.locator('[data-clan-save="clans/settings"]').click();
    await expect(box.locator('[data-clan-msg]')).toContainText('cooldowns.inviteDays');
    await box.locator('input[name="cooldowns.inviteDays"]').fill('۱۰');
    await box.locator('[data-clan-save="clans/settings"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    const s = await cln.settings();
    expect(s.roles.kick.includes('elder')).toBe(!orig.roles.kick.includes('elder'));
    expect(s.cooldowns.inviteDays).toBe(10);
    expect(s.xp.sources.length).toBe(before + 1);
    expect(s.xp.sources[before]).toMatchObject({ type: 'economy_event', event: 'e2e_win', amount: 7 });
    expect(s.create.cost).toEqual({});
    await cln.put('/settings', orig);
    await page.reload();
    await expect(root(page).locator('[data-clan-xp-source]')).toHaveCount(before);
  });

  test('leaderboardx: a clan board hides rewards and client submit, and the server refuses rewards on it', async ({ page }) => {
    await page.goto(CONSOLE_URL + '/#/leaderboardx/boards');
    const lroot = page.locator('#lbx-root');
    await expect(lroot.locator('[data-lbx-page="boards"]')).toBeVisible();
    await lroot.locator('[data-lbx-action="new-board"]').click();
    const p = page.locator('[data-lbx=panel] [role=dialog]');
    await expect(p.locator('select[name="scope"]')).toHaveValue('player');
    await expect(p.locator('[data-lbx-clan-note]')).toBeHidden();
    await p.locator('select[name="source.type"]').selectOption('client');
    await p.locator('select[name="scope"]').selectOption('clan');
    await expect(p.locator('[data-lbx-clan-note]')).toBeVisible();
    await expect(p.locator('[data-lbx-section="rewards"]')).toBeHidden();
    await expect(p.locator('select[name="source.type"]')).toHaveValue('economy_event');
    await expect(p.locator('input[name="limits.clientSubmit"]')).toBeDisabled();
    await p.locator('input[name="key"]').fill(BOARD);
    await p.locator('input[name="name-fa"]').fill('جدول کلن ' + short);
    await p.locator('input[name="source.event"]').fill('e2e_score');
    await p.locator('select[name="period"]').selectOption('none');
    await p.locator('[data-lbx-action="save-board"]').click();
    await expect(p).toHaveCount(0);
    await expect(lroot.locator(`[data-lbx-scope="${BOARD}"]`)).toHaveText('کلن');
    const b = (await lbx.get('/boards')).items.find((x: any) => x.board.key === BOARD).board;
    expect(b).toMatchObject({ scope: 'clan', rewards: [], limits: { clientSubmit: false } });
    const r = await lbx.raw('PUT', '/boards/' + b.id, { ...b, rewards: [{ fromRank: 1, toRank: 1, prizeKey: 'x' }] });
    expect(r.status).toBe(400);
    expect(r.data.error).toBe('LEADERBOARD_CLAN_NO_REWARDS');
    await lroot.locator(`tr[data-lbx-row="${BOARD}"]`).click();
    await expect(page.locator('[data-lbx=panel] select[name="scope"]')).toHaveValue('clan');
  });

  test('a clan is deleted from its panel after a confirm', async ({ page }) => {
    await openPage(page, 'clans/list');
    await root(page).locator('input[name="q"]').fill('e2e' + short);
    await root(page).locator(`tr[data-clan-row="${F.two}"]`).click();
    await panel(page).locator('[data-clan-action="delete-clan"]').click();
    await confirm(page);
    await expect(toasts(page).last()).toContainText('حذف شد');
    await expect(root(page).locator(`tr[data-clan-row="${F.two}"]`)).toHaveCount(0);
    expect((await cln.raw('GET', '/list/' + F.two)).status).toBe(404);
    F.two = '';
  });

  test('Ctrl K finds the clan pages; 390px English dark without sideways scroll', async ({ page, spawn }) => {
    await openPage(page, 'clans/list');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('سطح‌های کلن');
    await expect(page.locator('.pl-pal-row').first()).toContainText('سطح‌های کلن');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/clans\/levels/);
    await expect(root(page).locator('[data-clan-page="clans/levels"]')).toBeVisible();

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['list', 'Clans'], ['levels', 'Clan levels'], ['wars', 'Clan wars'], ['rounds', 'War rounds'], ['settings', 'Clan settings']]) {
      await m.goto(CONSOLE_URL + '/#/clans/' + sub);
      await expect(m.locator(`#clan-root [data-clan-page="clans/${sub}"] h1`)).toHaveText(title);
      await expect(m.getByText('Loading…')).toHaveCount(0);
      await noHorizontalScroll(m);
    }
    await m.goto(CONSOLE_URL + '/#/clans/list');
    await m.locator('#clan-root input[name="q"]').fill('e2e' + short);
    await m.locator(`#clan-root tr[data-clan-row="${F.one}"]`).click();
    await expect(m.locator('[data-clan=panel] [data-clan-section="members"]')).toContainText('Leader');
    await expect(m.locator('[data-clan=panel] [data-clan-section="xp"]')).toContainText('+50');
    await noHorizontalScroll(m);
    await m.keyboard.press('Escape');
    await m.goto(CONSOLE_URL + '/#/clans/rounds?configKey=' + WP);
    await m.locator('#clan-root [data-clan-list="rounds"] tbody tr').first().click();
    await expect(m.locator('[data-clan=panel] [data-clan-value="round-status"]')).toContainText('Finished');
    await noHorizontalScroll(m);
  });
});
