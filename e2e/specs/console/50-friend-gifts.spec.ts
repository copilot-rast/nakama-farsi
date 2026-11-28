import { expect, test } from '../../lib/fixtures';
import { API_URL, CONSOLE_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, noHorizontalScroll, open, panel, pickRef, root, rows, savePanel, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { users: [] as string[], items: [] as string[], types: [] as string[], tags: [] as string[] };
let savedRow = '';
let A: Player;
let B: Player;
let orb: any;
let gem: any;
let tag: any;
let giftId = '';

test.describe.configure({ mode: 'serial' });

async function befriend(a: Player, b: Player) {
  for (const [x, y] of [[a, b], [b, a]]) {
    const r = await fetch(API_URL + '/v2/friend?ids=' + encodeURIComponent(y.id), { method: 'POST', headers: { Authorization: 'Bearer ' + x.token } });
    expect(r.status).toBe(200);
  }
}

function restoreSettings(row: string) {
  if (!row) {
    sql('DELETE FROM economy_friend_gift_settings');
    return;
  }
  const i = row.lastIndexOf('|');
  const q = (v: string) => "'" + v.replace(/'/g, "''") + "'";
  sql('UPDATE economy_friend_gift_settings SET settings=' + q(row.slice(0, i)) + '::jsonb, updated_at=' + q(row.slice(i + 1)) + '::timestamptz');
}

function owns(user: string, itemId: string) {
  return sql(`SELECT count(*) FROM economy_user_item WHERE user_id='${user}' AND item_id='${itemId}'`) === '1';
}

test.beforeAll(async () => {
  savedRow = sql(`SELECT settings::text || '|' || updated_at::text FROM economy_friend_gift_settings`);
  const type = await base.post('/categories', { key: 'e2e_' + k + '_fct', keyType: 'item_type', name: { fa: 'نوع هدیهٔ کنسول' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  orb = await eco.post('/items', { key: 'e2e_' + k + '_orb', typeId: type.id, name: { fa: 'گوی ' + k.slice(-4), en: 'Orb' } });
  gem = await eco.post('/items', { key: 'e2e_' + k + '_gem', typeId: type.id, name: { fa: 'نگین ' + k.slice(-4), en: 'Gem' } });
  made.items.push(orb.id, gem.id);
  tag = await base.post('/tags', { key: 'e2e_' + k + '_fgv', keyType: 'user', name: { fa: 'هدیه‌دهندهٔ ویژه ' + k.slice(-4) }, description: {}, isPublic: false, metadata: {}, sort: 0 });
  made.tags.push(tag.id);
  A = await devicePlayer('e2e-fca-' + k, ('fca' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-fcb-' + k, ('fcb' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id);
  await befriend(A, B);
  await eco.post('/users/' + A.id + '/items', { itemId: orb.id });
  await eco.post('/users/' + A.id + '/items', { itemId: gem.id });
  await base.put('/players/' + A.id + '/tags', { tagIds: [tag.id] });
  const cur = (await eco.get('/friend-gift-settings')).settings;
  await eco.put('/friend-gift-settings', { ...cur, enabled: true, recipients: 'friends', minFriendDays: 0, minMutualFriends: 0, minSenderLevel: 0, minSenderAccountDays: 0, minRecipientAccountDays: 0,
    inventoryEnabled: true, itemsDefault: true, itemTypes: [], regift: 'block', delivery: 'instant', perSenderDaily: 5, perRecipientDaily: 10, perPairDaily: 0 });
});

test.afterAll(async () => {
  restoreSettings(savedRow);
  const users = made.users.map((x) => `'${x}'`).join(',');
  sql(`DELETE FROM economy_friend_gift WHERE sender_id IN (${users}) OR recipient_id IN (${users})`);
  sql(`DELETE FROM economy_friend_gift_rule WHERE name LIKE 'e2e_${k}%'`);
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_user_item WHERE item_id IN (${made.items.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
});

test.describe('friend gifts: hub, settings and rules', () => {
  test('the shop gifts tab has friend gifts with its rules, report and settings', async ({ page }) => {
    await open(page, '#/economy/friend-gifts');
    const sub = root(page).locator('[data-shop-sub="gifts"]');
    await expect(sub.locator('[data-shop-list]')).toHaveText(['کدهای هدیه', 'هدیه به دوست']);
    await expect(sub.locator('[data-shop-list="friend-gifts"]')).toHaveAttribute('aria-selected', 'true');
    await expect(sub.locator('[data-shop-tool]')).toHaveText(['قانون‌ها', 'گزارش', 'تنظیمات']);
  });

  test('settings save delivery, expiry, limits and currencies, and refuse a bad value', async ({ page }) => {
    await open(page, '#/economy/friend-gifts/settings');
    const form = page.locator('[data-settings-form="friend-gifts"]');
    await field(form, /^نحوهٔ رسیدن$/).locator('select').selectOption('accept');
    await field(form, 'مهلت قبول').locator('input').fill('۰');
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'مهلت قبول')).toContainText('عدد');
    await field(form, 'مهلت قبول').locator('input').fill('۴۸');
    await field(form, 'هدیه‌های فرستادهٔ هر بازیکن در روز').locator('input').fill('۳');
    await field(form, 'خرید با سکه که برگشت').locator('select').selectOption('sender');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    const st = (await eco.get('/friend-gift-settings')).settings;
    expect(st).toMatchObject({ delivery: 'accept', expiryHours: 48, perSenderDaily: 3, coinReturn: 'sender', recipients: 'friends', notifyChannel: 'in_app' });
    await eco.put('/friend-gift-settings', { ...st, delivery: 'instant', perSenderDaily: 5 });
  });

  test('a tag rule is made in its editor and listed with its limits', async ({ page }) => {
    await open(page, '#/economy/friend-gifts/rules/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^نام$/).locator('input').fill('e2e_' + k + ' ویژه');
    await pickRef(p, 'برچسب‌های بازیکن', 'e2e_' + k + '_fgv', /ویژه/);
    await field(p, 'هدیه‌های فرستادهٔ هر بازیکن در روز').locator('input').fill('۹');
    await savePanel(page);
    const list = await eco.get('/friend-gift-rules?q=' + encodeURIComponent('e2e_' + k));
    expect(list.items[0]).toMatchObject({ perSenderDaily: 9, perRecipientDaily: null, canSend: true, tagIds: [tag.id], players: 1 });
    await expect(rows(page).first().locator('[data-fg-rule-limits]')).toContainText('۹');
    const opts = await rpc(A, 'economy_friend_gift_options', { friendId: B.id });
    expect(opts.data.limits.sender.limit).toBe(9);
  });

  test('an item can be marked not giftable in its editor', async ({ page }) => {
    await open(page, '#/economy/items/' + gem.id);
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'هدیه به دوست').locator('select').selectOption('false');
    await savePanel(page);
    expect((await eco.get('/items/' + gem.id)).giftable).toBe(false);
    const res = await rpc(A, 'economy_friend_gift_send', { friendId: B.id, itemKey: 'e2e_' + k + '_gem' });
    expect(res.data.message).toBe('ITEM_NOT_GIFTABLE');
  });
});

test.describe('friend gifts: log, reverse, player card and report', () => {
  test('a sent gift is in the log with filters and opens to its detail', async ({ page }) => {
    const res = await rpc(A, 'economy_friend_gift_send', { friendId: B.id, itemKey: 'e2e_' + k + '_orb', message: 'برای تو' });
    expect(res.status).toBe(200);
    giftId = res.data.gift.id;
    await open(page, '#/economy/friend-gifts?userId=' + A.id);
    const row = rows(page).first();
    await expect(row.locator('[data-fg-what="e2e_' + k + '_orb"]')).toBeVisible();
    await expect(row.locator('[data-fg-status="delivered"]')).toContainText('رسیده');
    await expect(row).toContainText(A.username);
    await row.locator('td').first().click();
    await settle(page);
    const detail = root(page).locator(`[data-friend-gift="${giftId}"]`);
    await expect(detail).toContainText('برای تو');
    await expect(detail.locator('[data-fg-trail] li')).not.toHaveCount(0);
  });

  test('reversing a gift takes the item back to the sender', async ({ page }) => {
    await open(page, '#/economy/friend-gifts/' + giftId);
    await root(page).locator('[data-fg-reverse]').click();
    const dlg = page.locator('[data-economy="dialog"]');
    await expect(dlg.locator('[data-fg-return]')).toBeVisible();
    await dlg.locator('[data-fg-note]').fill('e2e chargeback');
    await dlg.locator('button', { hasText: 'برگرداندن هدیه' }).click();
    await expect(toasts(page).last()).toContainText('هدیه برگشت خورد');
    await expect(root(page).locator('[data-fg-status="reversed"]').first()).toBeVisible();
    expect(owns(A.id, orb.id)).toBe(true);
    expect(owns(B.id, orb.id)).toBe(false);
    const row = await eco.get('/friend-gifts/' + giftId);
    expect(row).toMatchObject({ status: 'reversed', outcome: 'returned', reverseNote: 'e2e chargeback' });
    await expect(root(page).locator('[data-fg-reverse]')).toHaveCount(0);
  });

  test('the player card counts gifts and links to the log', async ({ page }) => {
    await page.goto(CONSOLE_URL + '/#/players/' + A.id);
    const card = page.locator('[data-player-friend-gifts]');
    await expect(card.locator('h2')).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('[data-fg-counts]')).toContainText('۱ فرستاده');
    await expect(card.locator(`[data-fg-row="${giftId}"]`)).toContainText('برگشت‌خورده');
    await expect(card).toContainText('قانون: e2e_' + k);
    await expect(card.locator('[data-fg-log-link]')).toHaveAttribute('href', new RegExp('gifts/friends\\?userId=' + A.id));
  });

  test('the report counts the gift', async ({ page }) => {
    await open(page, '#/economy/friend-gifts/report?days=1');
    const stats = root(page).locator('[data-fg-stats]');
    await expect(stats).toContainText('هدیهٔ فرستاده');
    await expect(stats).toContainText('برگشت‌خورده');
  });

  test('the log, a gift and the settings fit a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await open(page, '#/economy/friend-gifts?userId=' + A.id);
    await expect(rows(page).first()).toBeVisible();
    await noHorizontalScroll(page);
    await open(page, '#/economy/friend-gifts/' + giftId);
    await expect(root(page).locator(`[data-friend-gift="${giftId}"]`)).toBeVisible();
    await noHorizontalScroll(page);
    await open(page, '#/economy/friend-gifts/settings');
    await noHorizontalScroll(page);
  });
});
