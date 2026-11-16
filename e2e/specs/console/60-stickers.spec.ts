import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { FIXTURES } from '../../lib/env';
import { base, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, pickRef, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { items: [] as string[], types: [] as string[] };
let savedSticker = '';
let savedQuota = '';
let item: any;

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

test.beforeAll(async () => {
  savedSticker = sql(`SELECT settings::text FROM base_sticker_settings WHERE id=1`);
  savedQuota = sql(`SELECT settings::text || '|' || updated_at::text FROM base_chat_quota_settings WHERE id=1`);
  const type = await base.post('/categories', { key: 'e2e_' + k + '_spt', keyType: 'item_type', name: { fa: 'نوع استیکر' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  item = await eco.post('/items', { key: 'e2e_' + k + '_spitem', typeId: type.id, name: { fa: 'کلید بستهٔ ویژه ' + k.slice(-4), en: 'Special pack key' } });
  made.items.push(item.id);
});

test.afterAll(async () => {
  if (savedSticker) sql(`UPDATE base_sticker_settings SET settings=${q(savedSticker)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_sticker_settings`);
  if (savedQuota) {
    const i = savedQuota.lastIndexOf('|');
    sql('UPDATE base_chat_quota_settings SET settings=' + q(savedQuota.slice(0, i)) + '::jsonb, updated_at=' + q(savedQuota.slice(i + 1)) + '::timestamptz WHERE id=1');
  }
  sql(`DELETE FROM base_sticker_pack WHERE key LIKE 'e2e_${k}%'`);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('stickers: hub, packs and settings', () => {
  test('the chat group has the stickers hub with packs and settings tabs, and the page search finds it', async ({ page }) => {
    await open(page, '#/moderation/players');
    await expect(page.locator('.x-drawer a[data-economy-page="stickers"]').first()).toHaveText('استیکرها');
    await open(page, '#/moderation/stickers');
    await expect(root(page).locator('[data-economy-tabs="stickers"] [data-economy-tab]')).toHaveText(['بسته‌ها', 'تنظیمات']);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input:visible').fill('استیکر');
    await expect(page.locator('.pl-pal-row').first()).toContainText('استیکرها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/moderation\/stickers/);
  });

  test('a free pack is made in the editor with an uploaded and a linked picture; an item pack needs its item', async ({ page }) => {
    await open(page, '#/moderation/stickers/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_happy');
    await fillTranslated(p, 'name', { fa: 'شاد ' + k.slice(-4), en: 'Happy' });
    await p.locator('[data-rows-add="stickers"]').click();
    await p.locator('[data-col="imageUrl"]').nth(0).locator('input[type="file"]').setInputFiles(path.join(FIXTURES, 'logo.svg'));
    await expect(p.locator('[data-col="imageUrl"]').nth(0).locator('input[dir="ltr"], input[type="text"]').first()).not.toHaveValue('');
    await p.locator('[data-rows-add="stickers"]').click();
    await p.locator('[data-col="imageUrl"]').nth(1).locator('input').first().fill('https://cdn.example.com/e2e/laugh.png');
    await p.locator('[data-col="key"]').nth(1).locator('input').fill('laugh');
    await savePanel(page);
    const list = await base.get('/stickers/packs?q=' + encodeURIComponent('e2e_' + k + '_happy'));
    expect(list.items[0]).toMatchObject({ key: 'e2e_' + k + '_happy', access: 'free', count: 2, isActive: true });
    expect(list.items[0].stickers.map((s: any) => s.key)).toEqual(['s1', 'laugh']);
    expect(list.items[0].stickers[0].imageUrl).not.toBe('');
    expect(list.items[0].stickers[1].imageUrl).toBe('https://cdn.example.com/e2e/laugh.png');
    await search(page, 'e2e_' + k + '_happy');
    await expect(rows(page).first().locator('[data-sp-access="free"]')).toContainText('رایگان برای همه');
    await expect(rows(page).first()).toContainText('۲ استیکر');

    await open(page, '#/moderation/stickers/new');
    await settlePanel(page);
    const v = panel(page);
    await field(v, 'کلید').locator('input').fill('e2e_' + k + '_vip');
    await fillTranslated(v, 'name', { fa: 'ویژه ' + k.slice(-4) });
    await field(v, 'دسترسی').locator('select').selectOption('item');
    await expect(field(v, 'آیتمی که بسته را باز می‌کند')).toBeVisible();
    await pickRef(v, 'آیتمی که بسته را باز می‌کند', item.key);
    await v.locator('[data-rows-add="stickers"]').click();
    await v.locator('[data-col="imageUrl"]').nth(0).locator('input').first().fill('/e2e/crown.png');
    await savePanel(page);
    const vip = (await base.get('/stickers/packs?q=' + encodeURIComponent('e2e_' + k + '_vip'))).items[0];
    expect(vip).toMatchObject({ access: 'item', itemKey: item.key, count: 1 });
    await search(page, 'e2e_' + k + '_vip');
    await expect(rows(page).first().locator('[data-sp-access="item"]')).toContainText('کلید بستهٔ ویژه');
  });

  test('sticker settings close private chats and keep stickers outside the quota', async ({ page }) => {
    await open(page, '#/moderation/stickers/settings');
    const form = root(page).locator('[data-settings-form="stickers"]');
    await expect(form.locator('label.x-toggle', { hasText: 'ارسال استیکر روشن باشد' })).toBeVisible();
    await form.locator('label.x-toggle', { hasText: 'پیام خصوصی' }).click();
    await form.locator('label.x-toggle', { hasText: 'هر استیکر یک پیام از سهمیه کم کند' }).click();
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    const st = await base.get('/stickers/settings');
    expect(st).toMatchObject({ enabled: true, direct: false, countQuota: false, groups: true, rooms: true });
  });

  test('the chat quota settings can stop counting private messages', async ({ page }) => {
    await base.put('/chat-quota/settings', { enabled: true, dailyMessages: 5 });
    await open(page, '#/moderation/chat-quota/settings');
    const form = root(page).locator('[data-settings-form="chat-quota"]');
    await form.locator('label.x-toggle', { hasText: 'پیام خصوصی شمرده شود' }).click();
    await form.locator('label.x-toggle', { hasText: 'اتاق' }).first().click();
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    expect(await base.get('/chat-quota/settings')).toMatchObject({ countDirect: false, countRooms: true });
  });

  test('the hub reads in English', async ({ page }) => {
    await open(page, '#/moderation/stickers');
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(root(page).locator('[data-economy-tabs="stickers"] [data-economy-tab]')).toHaveText(['Packs', 'Settings']);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('packs, settings and the editor fit a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const hash of ['#/moderation/stickers', '#/moderation/stickers/settings', '#/moderation/stickers/new']) {
      await open(page, hash);
      await settle(page);
      await test.step(hash, () => noHorizontalScroll(page));
    }
  });
});
