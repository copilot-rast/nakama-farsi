import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, pickRef, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { prizes: [] as string[], users: [] as string[], metas: [] as string[], currencies: [] as string[], tags: [] as string[] };
let saved = '';
let A: Player;
let B: Player;
let hair = '';
let shirt = '';
let female = '';
let xp: any;
let gem: any;
let vip: any;

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

function code(res: any) {
  return res.data && res.data.message;
}

function itemsOf(cat: any) {
  const out: Record<string, any> = {};
  for (const c of cat.categories) for (const it of c.items) out[c.key + '/' + it.key] = it;
  return out;
}

test.beforeAll(async () => {
  saved = sql(`SELECT settings::text FROM economy_avatar_settings WHERE id=1`);
  gem = await eco.post('/currencies', { key: 'e2e_' + k + '_gem', name: { fa: 'جم ' + k.slice(-4), en: 'Gem' }, type: 'soft' });
  made.currencies.push(gem.id);
  xp = await eco.post('/metas', { key: 'e2e_' + k + '_xp', name: { fa: 'تجربه', en: 'XP' }, defaultValue: 0 });
  made.metas.push(xp.id);
  vip = await base.post('/tags', { key: 'e2e_' + k + '_vip', keyType: 'user', name: { fa: 'ویژه' }, description: {}, sort: 0, metadata: {} });
  made.tags.push(vip.id);
  await eco.post('/avatar-types', { key: 'e2e_' + k + '_male', name: { fa: 'پسر' }, sort: -100000 });
  A = await devicePlayer('e2e-avd-a-' + k, ('avda' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-avd-b-' + k, ('avdb' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id);
  await eco.put('/avatar-settings', { enabled: true, maxAvatars: 1, nameMaxLength: 24, autoCreate: true, keepLast: true, slotFull: 'replace', keepBought: true, tagLimits: [] });
});

test.afterAll(async () => {
  if (saved) sql(`UPDATE economy_avatar_settings SET settings=${q(saved)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM economy_avatar_settings`);
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_avatar_category WHERE key LIKE 'e2e_${k}%'`);
  sql(`DELETE FROM economy_avatar_type WHERE key LIKE 'e2e_${k}%'`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (${made.prizes.map((x) => `'${x}'`).join(',') || 'NULL'})`);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
  for (const id of made.currencies) await eco.del('/currencies/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
});

test.describe('avatar designer: console and game', () => {
  test('the shop has the avatar tab with categories, items, types and settings', async ({ page }) => {
    await open(page, '#/economy/items');
    await expect(root(page).locator('[data-shop-tab="avatar"]')).toHaveText('آواتار');
    await open(page, '#/economy/avatar');
    await expect(root(page).locator('[data-shop-sub="avatar"] [data-shop-list]')).toHaveText(['دسته‌ها', 'آیتم‌ها', 'نوع‌ها']);
    await expect(root(page).locator('[data-shop-sub="avatar"] [data-shop-tool]')).toHaveText(['تنظیمات']);
  });

  test('a type, two categories and a paid item are made in the editors', async ({ page }) => {
    await open(page, '#/economy/avatar/types/new');
    await settlePanel(page);
    let p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_female');
    await fillTranslated(p, 'name', { fa: 'دختر ' + k.slice(-4), en: 'Female' });
    await savePanel(page);
    female = (await eco.get('/avatar-types?q=' + encodeURIComponent('e2e_' + k + '_female'))).items[0].id;

    await open(page, '#/economy/avatar/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_shirt');
    await fillTranslated(p, 'name', { fa: 'پیراهن ' + k.slice(-4), en: 'Shirt' });
    await field(p, 'لایه').locator('input').fill('۲');
    await savePanel(page);
    shirt = (await eco.get('/avatar-categories?q=' + encodeURIComponent('e2e_' + k + '_shirt'))).items[0].id;
    hair = (await eco.post('/avatar-categories', { key: 'e2e_' + k + '_hair', name: { fa: 'مو' }, layer: 3 })).id;

    await open(page, '#/economy/avatar/items/new');
    await settlePanel(page);
    p = panel(page);
    await pickRef(p, 'دسته', 'پیراهن ' + k.slice(-4));
    await field(p, 'کلید').locator('input').fill('red');
    await fillTranslated(p, 'name', { fa: 'قرمز', en: 'Red' });
    await field(p, 'مقدار (مثلاً رنگ)').locator('input').fill('#E53935');
    await field(p, 'این آیتم چطور باز می‌شود').locator('select').selectOption('currency');
    await pickRef(p, 'ارز قیمت', 'جم ' + k.slice(-4));
    await field(p, /^قیمت$/).locator('input').fill('۴۰');
    await savePanel(page);
    const red = (await eco.get('/avatar-items?categoryId=' + shirt)).items[0];
    expect(red).toMatchObject({ key: 'red', access: 'currency', price: { currencyKey: gem.key, amount: 40 }, value: '#E53935' });
    await open(page, '#/economy/avatar/items');
    await search(page, 'red');
    await expect(rows(page).first()).toContainText('۴۰');
  });

  test('items are free, paid, locked by meta and limited to one avatar type', async () => {
    await eco.post('/avatar-items', { categoryId: hair, key: 'short', name: { fa: 'کوتاه' }, value: '#111', isDefault: true });
    await eco.post('/avatar-items', { categoryId: hair, key: 'braid', name: { fa: 'بافته' }, value: '#222', typeIds: [female] });
    await eco.post('/avatar-items', { categoryId: shirt, key: 'star', name: { fa: 'ستاره' }, value: '#333', requirements: { metas: [{ metaId: xp.id, operator: 'gte', value: 300 }] } });
    const cat = await rpc(A, 'economy_avatar_catalog', { lang: 'fa' });
    expect(cat.status).toBe(200);
    const items = itemsOf(cat.data);
    expect(items['e2e_' + k + '_hair/short']).toMatchObject({ access: 'free', usable: true, equipped: true });
    expect(items['e2e_' + k + '_shirt/red']).toMatchObject({ access: 'currency', owned: false, price: { amount: 40 } });
    expect(items['e2e_' + k + '_shirt/star']).toMatchObject({ locked: true, usable: false, missing: [{ type: 'meta', key: xp.key, value: 300 }] });

    expect(code(await rpc(A, 'economy_avatar_equip', { category: 'e2e_' + k + '_hair', key: 'braid' }))).toBe('AVATAR_TYPE_NOT_ALLOWED');
    expect(code(await rpc(A, 'economy_avatar_equip', { category: 'e2e_' + k + '_shirt', key: 'star' }))).toBe('AVATAR_ITEM_LOCKED');
    expect(code(await rpc(A, 'economy_avatar_buy', { category: 'e2e_' + k + '_shirt', key: 'red' }))).toBe('INSUFFICIENT_BALANCE');
    const gems = await eco.post('/prizes', { key: 'e2e_' + k + '_gems', name: { fa: 'جم' }, isRepeatable: true, items: [{ type: 'currency', key: gem.key, action: 'add', value: 100 }] });
    made.prizes.push(gems.id);
    await eco.post('/users/' + A.id + '/grant', { prizeId: gems.id });
    const buy = await rpc(A, 'economy_avatar_buy', { category: 'e2e_' + k + '_shirt', key: 'red', requestId: 'buy-1' });
    expect(buy.status).toBe(200);
    expect(buy.data.item).toMatchObject({ owned: true, usable: true });
    expect(buy.data.wallet[gem.key]).toBe(60);
    const wear = await rpc(A, 'economy_avatar_equip', { category: 'e2e_' + k + '_shirt', key: 'red' });
    expect(wear.data.avatar.look.items['e2e_' + k + '_shirt']).toEqual(['red']);

    sql(`INSERT INTO economy_user_meta (user_id, meta_id, value) VALUES ('${A.id}','${xp.id}',300) ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value`);
    const star = await rpc(A, 'economy_avatar_equip', { category: 'e2e_' + k + '_shirt', key: 'star' });
    expect(star.status).toBe(200);
    expect(star.data.avatar.look.items['e2e_' + k + '_shirt']).toEqual(['star']);
  });

  test('a VIP gets a second avatar with its own items and the others see the active one', async () => {
    expect(code(await rpc(A, 'economy_avatar_create', { name: 'دوم', type: 'e2e_' + k + '_female' }))).toBe('AVATAR_LIMIT');
    await eco.put('/avatar-settings', { enabled: true, maxAvatars: 1, nameMaxLength: 24, autoCreate: true, keepLast: true, slotFull: 'replace', keepBought: true, tagLimits: [{ tagId: vip.id, maxAvatars: 2 }] });
    sql(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ('${A.id}','${vip.id}') ON CONFLICT DO NOTHING`);
    const made2 = await rpc(A, 'economy_avatar_create', { name: 'سارا', type: 'e2e_' + k + '_female', activate: true });
    expect(made2.status).toBe(200);
    expect(made2.data.limit).toMatchObject({ max: 2, count: 2, canCreate: false });
    const second = made2.data.avatar.id;
    const braid = await rpc(A, 'economy_avatar_equip', { id: second, category: 'e2e_' + k + '_hair', key: 'braid' });
    expect(braid.data.avatar.look.items['e2e_' + k + '_hair']).toEqual(['braid']);
    const list = await rpc(A, 'economy_avatar_list', {});
    const first = list.data.avatars.find((a: any) => a.id !== second);
    expect(first.look.items['e2e_' + k + '_hair']).toEqual(['short']);
    expect(first.look.items['e2e_' + k + '_shirt']).toEqual(['star']);
    const looks = await rpc(B, 'economy_avatar_looks', { userIds: [A.id, 'bot'] });
    expect(looks.data.avatars[A.id].look.items['e2e_' + k + '_hair']).toEqual(['braid']);
    expect(code(await rpc(A, 'economy_avatar_create', { name: 'سوم' }))).toBe('AVATAR_LIMIT');
    const back = await rpc(A, 'economy_avatar_set_active', { id: first.id });
    expect(back.data.activeId).toBe(first.id);
    const old = await rpc(B, 'base_avatar_get', { userIds: [A.id] });
    expect(old.data.looks[A.id].parts['e2e_' + k + '_hair']).toBe('short');
  });

  test('the settings page changes the avatar count', async ({ page }) => {
    await open(page, '#/economy/avatar/settings');
    const form = page.locator('[data-settings-form="avatar"]');
    await field(form, 'بیشترین تعداد آواتار').first().locator('input').fill('۳');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    expect(await eco.get('/avatar-settings')).toMatchObject({ maxAvatars: 3, tagLimits: [{ tagId: vip.id, maxAvatars: 2 }] });
    expect((await rpc(B, 'economy_avatar_list', {})).data.limit.max).toBe(3);
    await form.locator('label.x-toggle', { hasText: 'آواتارساز روشن باشد' }).click();
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    expect(code(await rpc(A, 'economy_avatar_equip', { category: 'e2e_' + k + '_hair', key: 'short' }))).toBe('AVATAR_DISABLED');
    await eco.put('/avatar-settings', { enabled: true, maxAvatars: 1, nameMaxLength: 24, autoCreate: true, keepLast: true, slotFull: 'replace', keepBought: true, tagLimits: [] });
  });

  test('the hub reads in English', async ({ page }) => {
    await open(page, '#/economy/avatar');
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(root(page).locator('[data-shop-sub="avatar"] [data-shop-list]')).toHaveText(['Categories', 'Items', 'Types']);
    await expect(root(page).locator('[data-shop-sub="avatar"] [data-shop-tool]')).toHaveText(['Settings']);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('every tab and editor fits a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const hash of ['#/economy/avatar', '#/economy/avatar/items', '#/economy/avatar/types', '#/economy/avatar/settings', '#/economy/avatar/new', '#/economy/avatar/items/new']) {
      await open(page, hash);
      await settle(page);
      await test.step(hash, () => noHorizontalScroll(page));
    }
  });
});
