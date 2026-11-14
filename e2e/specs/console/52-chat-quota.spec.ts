import { expect, test } from '../../lib/fixtures';
import { CONSOLE_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { Chat, notice } from '../../lib/chat';
import { sql } from '../../lib/db';
import { field, noHorizontalScroll, open, panel, pickRef, root, rows, savePanel, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { users: [] as string[], metas: [] as string[], prizes: [] as string[], products: [] as string[], tags: [] as string[] };
const sockets: Chat[] = [];
let savedRow = '';
let packKey = '';
let tag: any;
let A: Player;
let B: Player;

test.describe.configure({ mode: 'serial' });

function restoreSettings(row: string) {
  const i = row.lastIndexOf('|');
  const q = (v: string) => "'" + v.replace(/'/g, "''") + "'";
  sql('UPDATE base_chat_quota_settings SET settings=' + q(row.slice(0, i)) + '::jsonb, updated_at=' + q(row.slice(i + 1)) + '::timestamptz WHERE id=1');
}

async function dm(from: Player, to: Player, n: number) {
  const c = await Chat.open(from);
  sockets.push(c);
  await c.joinDirect(to.id);
  const out: boolean[] = [];
  for (let i = 0; i < n; i++) out.push((await c.send('پیام ' + i)).ok);
  return out;
}

test.beforeAll(async () => {
  savedRow = sql(`SELECT settings::text || '|' || updated_at::text FROM base_chat_quota_settings WHERE id=1`);
  packKey = 'e2e_' + k + '_cqpack';
  const meta = await eco.post('/metas', { key: packKey, name: { fa: 'بستهٔ پیام ' + k.slice(-4), en: 'Message pack' }, defaultValue: 0, serverOnly: false, settings: {} });
  made.metas.push(meta.id);
  const prize = await eco.post('/prizes', { key: 'e2e_' + k + '_cqprize', name: { fa: '۲۰ پیام' }, isRepeatable: true, items: [{ type: 'meta', key: packKey, action: 'add', value: 20 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  const product = await eco.post('/products', { key: 'e2e_' + k + '_cq20', isActive: true, name: { fa: 'بستهٔ ۲۰ پیام ' + k.slice(-4), en: '20 messages' }, prizeId: prize.id, paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: 5 } });
  made.products.push(product.id);
  tag = await base.post('/tags', { key: 'e2e_' + k + '_cqv', keyType: 'user', name: { fa: 'گفتگوی ویژه ' + k.slice(-4) }, description: {}, isPublic: false, metadata: {}, sort: 0 });
  made.tags.push(tag.id);
  A = await devicePlayer('e2e-cqa-' + k, ('cqa' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-cqb-' + k, ('cqb' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id);
  await updateAccount(A, { lang_tag: 'fa' });
  await base.put('/chat-quota/settings', { enabled: false });
});

test.afterAll(async () => {
  for (const s of sockets) s.close();
  if (savedRow) restoreSettings(savedRow);
  sql(`DELETE FROM base_chat_quota_rule WHERE name LIKE 'e2e_${k}%'`);
  for (const id of made.users) await deleteUser(id);
  for (const id of made.products) await eco.del('/products/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
});

test.describe('private chat quota: hub, settings and rules', () => {
  test('the chat group has the private chat quota hub with rules, settings and report tabs', async ({ page }) => {
    await open(page, '#/moderation/players');
    const link = page.locator('.x-drawer a[data-economy-page="chat-quota"]').first();
    await expect(link).toHaveText('سهمیهٔ پیام خصوصی');
    await open(page, '#/moderation/chat-quota');
    await expect(root(page).locator('[data-economy-tabs="chat-quota"] [data-economy-tab]')).toHaveText(['قانون‌ها', 'تنظیمات', 'آمار']);
  });

  test('settings turn the quota on with a daily count, a pack meta and a reset time, refuse a bad time and show the pack products', async ({ page }) => {
    await open(page, '#/moderation/chat-quota/settings');
    const form = root(page).locator('[data-settings-form="chat-quota"]');
    await expect(form.locator('[data-cq-summary]')).toBeVisible();
    await form.locator('label.x-toggle', { hasText: 'سهمیهٔ پیام خصوصی روشن باشد' }).click();
    await field(form, 'پیام رایگان در روز').locator('input').fill('۲');
    await field(form, 'ساعت شروع روز تازه').locator('input').fill('25:00');
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'ساعت شروع روز تازه')).toContainText('۰۰:۰۰ تا ۲۳:۵۹');
    await field(form, 'ساعت شروع روز تازه').locator('input').fill('۰۰:۰۰');
    await pickRef(form, 'متای بستهٔ پیام', packKey);
    await field(form, 'ترتیب مصرف').locator('select').selectOption('free_first');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    const st = await base.get('/chat-quota/settings');
    expect(st).toMatchObject({ enabled: true, dailyMessages: 2, packMeta: packKey, resetTime: '00:00', timeZone: 'Asia/Tehran', countMode: 'messages', order: 'free_first', replyFree: 'off', friendPolicy: 'same' });
    expect(st.detectedProducts.map((p: any) => p.key)).toContain('e2e_' + k + '_cq20');
    await expect(form.locator('[data-cq-summary]')).toContainText('۲ پیام رایگان در روز');
    await expect(form.locator(`[data-cq-product="e2e_${k}_cq20"]`)).toContainText('+۲۰ پیام');

    await field(form, 'شمارش').locator('select').selectOption('both');
    await expect(field(form, 'گفتگوی تازه در روز')).toBeVisible();
    await field(form, 'شمارش').locator('select').selectOption('messages');
    await expect(field(form, 'گفتگوی تازه در روز')).toBeHidden();
  });

  test('a tag rule is made unlimited in its editor and listed', async ({ page }) => {
    await open(page, '#/moderation/chat-quota/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'نام قانون').locator('input').fill('e2e_' + k + ' ویژه');
    await pickRef(p, 'برچسب‌های بازیکن', 'e2e_' + k + '_cqv', /گفتگوی ویژه/);
    await p.locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(p).toContainText('دست‌کم یک سقف بگذارید');
    await p.locator('label.x-toggle', { hasText: 'پیام خصوصی بی‌سقف' }).click();
    await expect(field(p, 'پیام رایگان در روز')).toBeHidden();
    await savePanel(page);
    const list = await base.get('/chat-quota/rules?q=' + encodeURIComponent('e2e_' + k));
    expect(list.items[0]).toMatchObject({ unlimited: true, dailyMessages: null, tagIds: [tag.id], isActive: true });
    await expect(rows(page).first().locator('[data-cq-limits]')).toContainText('بی‌سقف');
  });
});

test.describe('private chat quota: player card and report', () => {
  test('the player card shows today, gives extra messages, takes a pack back and resets today', async ({ page }) => {
    expect(await dm(A, B, 3)).toEqual([true, true, false]);
    await page.goto(CONSOLE_URL + '/#/players/' + A.id);
    const card = page.locator('[data-player-chat-quota]');
    await expect(card.locator('h2')).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('[data-cq-free]')).toContainText('۰');
    await expect(card.locator('[data-cq-free]')).toContainText('۲ از ۲ مصرف شده');
    await expect(card.locator('[data-cq-tags]')).toContainText('۱ رد شد');
    await expect(card.locator('[data-cq-today]')).toContainText('۲ پیام');

    await card.locator('[data-cq-action="grant"]').click();
    const dlg = page.locator('[data-form-dialog="chat-quota-grant"]');
    await dlg.locator('[data-cq-grant="amount"]').fill('۵');
    await dlg.locator('[data-cq-grant="note"]').fill('e2e جبران');
    await dlg.locator('[data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('پیام اضافه داده شد');
    await expect(card.locator('[data-cq-packs]')).toContainText('۵');
    await expect(card.locator('[data-cq-lots]')).toContainText('e2e جبران');
    expect(await dm(A, B, 2)).toEqual([true, true]);
    expect((await rpc(A, 'base_chat_quota', {})).data).toMatchObject({ packsLeft: 3, today: { fromPacks: 2 } });

    await page.reload();
    await expect(card.locator('[data-cq-packs]')).toContainText('۳', { timeout: 20_000 });
    await card.locator('[data-cq-revoke]').first().click();
    await page.locator('[data-form-dialog="chat-quota-revoke"] [data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('بسته پس گرفته شد');
    await expect(card.locator('[data-cq-packs]')).toContainText('۰');

    await card.locator('[data-cq-action="reset"]').click();
    await page.locator('[data-form-dialog="chat-quota-reset"] [data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('سهمیهٔ امروز صفر شد');
    await expect(card.locator('[data-cq-free]')).toContainText('۰ از ۲ مصرف شده');
    expect(await dm(A, B, 1)).toEqual([true]);
  });

  test('a player with the tag is unlimited on the card', async ({ page }) => {
    await base.put('/players/' + B.id + '/tags', { tagIds: [tag.id] });
    expect(await dm(B, A, 4)).toEqual([true, true, true, true]);
    await page.goto(CONSOLE_URL + '/#/players/' + B.id);
    const card = page.locator('[data-player-chat-quota]');
    await expect(card.locator('[data-cq-tags]')).toContainText('بی‌سقف', { timeout: 20_000 });
    await expect(card.locator('[data-cq-tags]')).toContainText('قانون: e2e_' + k);
    await expect(card.locator('[data-cq-free]')).toContainText('∞');
  });

  test('the report counts refused messages, pack messages and console grants', async ({ page }) => {
    const c = await Chat.open(A);
    sockets.push(c);
    await c.joinDirect(B.id);
    await c.send('x');
    const refused = await c.send('y');
    expect(notice((refused as any).message)).toMatchObject({ code: 'DM_QUOTA_EXCEEDED' });
    await open(page, '#/moderation/chat-quota/report?days=1');
    const stats = root(page).locator('[data-cq-stats]');
    await expect(stats).toContainText('ردشده به خاطر سهمیه');
    await expect(stats).toContainText('از بسته‌ها');
    await expect(root(page).locator('[data-cq-granted-admin]')).not.toHaveText('۰');
    await expect(root(page).locator('[data-cq-daily]')).toBeVisible();
    await expect(root(page).locator('a[href="#/players/' + A.id + '"]')).toBeVisible();
    const rep = await base.get('/chat-quota/report?days=1');
    expect(rep.totals.blocked).toBeGreaterThanOrEqual(2);
    expect(rep.totals.fromPacks).toBeGreaterThanOrEqual(2);
    expect(rep.packs.granted.admin).toBeGreaterThanOrEqual(5);
  });

  test('the settings and tabs read in English', async ({ page }) => {
    await open(page, '#/moderation/chat-quota/settings');
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(root(page).locator('[data-settings-form="chat-quota"]')).toContainText('Private chat quota is on');
    await expect(root(page).locator('[data-economy-tabs="chat-quota"] [data-economy-tab]')).toHaveText(['Rules', 'Settings', 'Report']);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('settings, rules, report and the player card fit a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const hash of ['#/moderation/chat-quota/settings', '#/moderation/chat-quota', '#/moderation/chat-quota/report']) {
      await open(page, hash);
      await settle(page);
      await noHorizontalScroll(page);
    }
    await page.goto(CONSOLE_URL + '/#/players/' + A.id);
    await expect(page.locator('[data-player-chat-quota] h2')).toBeVisible({ timeout: 20_000 });
    await noHorizontalScroll(page);
  });
});
