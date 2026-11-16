import path from 'node:path';
import type { Locator } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { FIXTURES } from '../../lib/env';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { items: [] as string[], types: [] as string[], users: [] as string[], parts: [] as string[] };
let saved = '';
let item: any;
let A: Player;
let B: Player;

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

async function pickIn(scope: Locator, text: string) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

function err(res: any) {
  try {
    return JSON.parse(res.data.message);
  } catch {
    return { code: res.data && res.data.message };
  }
}

test.beforeAll(async () => {
  saved = sql(`SELECT settings::text FROM base_avatar_settings WHERE id=1`);
  const type = await base.post('/categories', { key: 'e2e_' + k + '_avt', keyType: 'item_type', name: { fa: 'کلاه آواتار' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  item = await eco.post('/items', { key: 'e2e_' + k + '_crown', typeId: type.id, name: { fa: 'تاج ' + k.slice(-4), en: 'Crown' } });
  made.items.push(item.id);
  A = await devicePlayer('e2e-av-a-' + k, ('ava' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-av-b-' + k, ('avb' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id);
  await base.put('/avatar/settings', { enabled: true });
});

test.afterAll(async () => {
  if (saved) sql(`UPDATE base_avatar_settings SET settings=${q(saved)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_avatar_settings`);
  sql(`DELETE FROM base_avatar_part WHERE key LIKE 'e2e_${k}%'`);
  for (const id of made.users) await deleteUser(id);
  if (made.items.length) sql(`DELETE FROM economy_user_item WHERE item_id IN (${made.items.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('avatar builder: console and game', () => {
  test('the sales group has the avatar builder hub with parts and settings tabs', async ({ page }) => {
    await open(page, '#/economy/items');
    await expect(page.locator('.x-drawer a[data-economy-page="avatar"]').first()).toHaveText('آواتارساز');
    await open(page, '#/economy/avatar');
    await expect(root(page).locator('[data-economy-tabs="avatar"] [data-economy-tab]')).toHaveText(['بخش‌ها', 'تنظیمات']);
  });

  test('a required skin part with colour values and a hat part with an uploaded free hat and an item hat are made in the editor', async ({ page }) => {
    await open(page, '#/economy/avatar/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_skin');
    await fillTranslated(p, 'name', { fa: 'رنگ پوست ' + k.slice(-4), en: 'Skin' });
    await p.locator('label.x-toggle', { hasText: 'بازیکن حتماً یک گزینه داشته باشد' }).click();
    await p.locator('[data-rows-add="options"]').click();
    await p.locator('[data-col="key"]').nth(0).locator('input').fill('light');
    await p.locator('[data-col="value"]').nth(0).locator('input').fill('#F1C27D');
    await p.locator('[data-rows-add="options"]').click();
    await p.locator('[data-col="key"]').nth(1).locator('input').fill('dark');
    await p.locator('[data-col="value"]').nth(1).locator('input').fill('#8D5524');
    await savePanel(page);
    const skin = (await base.get('/avatar/parts?q=' + encodeURIComponent('e2e_' + k + '_skin'))).items[0];
    expect(skin).toMatchObject({ required: true, count: 2 });
    expect(skin.options.map((o: any) => [o.key, o.value, o.isDefault])).toEqual([['light', '#F1C27D', true], ['dark', '#8D5524', false]]);
    made.parts.push(skin.id);

    await open(page, '#/economy/avatar/new');
    await settlePanel(page);
    const h = panel(page);
    await field(h, 'کلید').locator('input').fill('e2e_' + k + '_hat');
    await fillTranslated(h, 'name', { fa: 'کلاه ' + k.slice(-4) });
    await field(h, 'لایه').locator('input').fill('۵');
    await h.locator('[data-rows-add="options"]').click();
    await h.locator('[data-col="imageUrl"]').nth(0).locator('input[type="file"]').setInputFiles(path.join(FIXTURES, 'logo.svg'));
    await h.locator('[data-col="key"]').nth(0).locator('input').fill('cap');
    await h.locator('[data-col="isDefault"]').nth(0).locator('[role="checkbox"]').first().evaluate((el) => { if (el.getAttribute('aria-checked') === 'true') (el as HTMLElement).click(); });
    await h.locator('[data-rows-add="options"]').click();
    await h.locator('[data-col="imageUrl"]').nth(1).locator('input').first().fill('/e2e/crown.png');
    await h.locator('[data-col="key"]').nth(1).locator('input').fill('crown');
    await h.locator('[data-col="access"]').nth(1).locator('select').selectOption('item');
    await h.locator('[data-col="isDefault"]').nth(1).locator('[role="checkbox"]').first().evaluate((el) => { if (el.getAttribute('aria-checked') === 'true') (el as HTMLElement).click(); });
    await pickIn(h.locator('[data-col="itemKey"]').nth(1), item.key);
    await savePanel(page);
    const hat = (await base.get('/avatar/parts?q=' + encodeURIComponent('e2e_' + k + '_hat'))).items[0];
    expect(hat).toMatchObject({ layer: 5, required: false, count: 2 });
    expect(hat.options[1]).toMatchObject({ key: 'crown', access: 'item', itemKey: item.key });
    made.parts.push(hat.id);
    await search(page, 'e2e_' + k + '_hat');
    await expect(rows(page).first()).toContainText('۲ گزینه');
  });

  test('the game reads the parts, saves a look, cannot wear a locked hat until it owns the item, and sees other players', async () => {
    const res = await rpc(A, 'base_avatar', { lang: 'fa' });
    expect(res.status).toBe(200);
    const parts = (res.data.parts as any[]).filter((p) => p.key.startsWith('e2e_' + k));
    expect(parts.map((p) => p.key)).toEqual(['e2e_' + k + '_hat', 'e2e_' + k + '_skin']);
    const crown = parts[0].options.find((o: any) => o.key === 'crown');
    expect(crown).toMatchObject({ access: 'item', owned: false, itemKey: item.key, item: { key: item.key } });
    expect(res.data.look.parts['e2e_' + k + '_skin']).toBe('light');
    expect(res.data.look.parts['e2e_' + k + '_hat']).toBeUndefined();

    const locked = await rpc(A, 'base_avatar_save', { look: { ['e2e_' + k + '_skin']: 'dark', ['e2e_' + k + '_hat']: 'crown' } });
    expect(err(locked)).toMatchObject({ code: 'AVATAR_OPTION_LOCKED', meta: { option: 'crown', itemKey: item.key } });
    await eco.post('/users/' + A.id + '/items', { itemId: item.id });
    const ok = await rpc(A, 'base_avatar_save', { look: { ['e2e_' + k + '_skin']: 'dark', ['e2e_' + k + '_hat']: 'crown' } });
    expect(ok.status).toBe(200);
    const layers = (ok.data.look.layers as any[]).filter((l) => l.part.startsWith('e2e_' + k));
    expect(layers).toEqual([
      { part: 'e2e_' + k + '_skin', option: 'dark', imageUrl: '', value: '#8D5524', layer: 0 },
      { part: 'e2e_' + k + '_hat', option: 'crown', imageUrl: '/e2e/crown.png', value: '', layer: 5 },
    ]);

    const others = await rpc(B, 'base_avatar_get', { userIds: [A.id, 'bot'] });
    expect(others.data.looks[A.id].parts['e2e_' + k + '_hat']).toBe('crown');
    expect(others.data.looks.bot.parts['e2e_' + k + '_skin']).toBe('light');
    expect(err(await rpc(A, 'base_avatar_save', { look: { ['e2e_' + k + '_skin']: '' } })).code).toBe('AVATAR_PART_REQUIRED');
  });

  test('turning the builder off empties it for players', async ({ page }) => {
    await open(page, '#/economy/avatar/settings');
    const form = root(page).locator('[data-settings-form="avatar"]');
    await form.locator('label.x-toggle', { hasText: 'آواتارساز روشن باشد' }).click();
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    expect(await base.get('/avatar/settings')).toMatchObject({ enabled: false });
    expect((await rpc(A, 'base_avatar', {})).data).toMatchObject({ enabled: false, parts: [] });
    expect(err(await rpc(A, 'base_avatar_save', { look: {} })).code).toBe('AVATAR_DISABLED');
    await base.put('/avatar/settings', { enabled: true });
  });

  test('the hub reads in English', async ({ page }) => {
    await open(page, '#/economy/avatar');
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(root(page).locator('[data-economy-tabs="avatar"] [data-economy-tab]')).toHaveText(['Parts', 'Settings']);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('parts, settings and the editor fit a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const hash of ['#/economy/avatar', '#/economy/avatar/settings', '#/economy/avatar/new']) {
      await open(page, hash);
      await settle(page);
      await test.step(hash, () => noHorizontalScroll(page));
    }
  });
});
