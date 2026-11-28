import { expect, test, type Locator } from '../../lib/fixtures';
import { base, eco, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, pickRef, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { items: [] as string[], types: [] as string[], tags: [] as string[], prizes: [] as string[], games: [] as string[] };
let savedSettings = '';
let item: any;
let vip: any;

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

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

test.beforeAll(async () => {
  savedSettings = sql(`SELECT settings::text FROM economy_tag_discount_settings WHERE id=1`);
  const type = await base.post('/categories', { key: k + '_tdt', keyType: 'item_type', name: { fa: 'نوع تخفیفی' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  item = await eco.post('/items', { key: k + '_tditem', typeId: type.id, name: { fa: 'بستهٔ استیکر ' + k.slice(-4), en: 'Sticker pack' }, price: { currencyKey: 'coin', amount: 200 } });
  made.items.push(item.id);
  vip = await base.post('/tags', { key: k + '_vip', keyType: 'user', name: { fa: 'ویژه ' + k.slice(-4), en: 'VIP' }, description: {}, sort: 0, metadata: {} });
  made.tags.push(vip.id);
});

test.afterAll(async () => {
  if (savedSettings) sql(`UPDATE economy_tag_discount_settings SET settings=${q(savedSettings)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM economy_tag_discount_settings`);
  sql(`DELETE FROM economy_tag_discount WHERE key LIKE '${k}%'`);
  for (const id of made.games) await eco.del('/chance-games/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('tag perks: timed tag prizes, tag discounts and chests', () => {
  test('the shop has the tag discounts tab with its tools, and Ctrl K finds it', async ({ page }) => {
    await open(page, '#/economy/tag-discounts');
    await expect(root(page).locator('[data-shop-tab="discounts"]')).toHaveText('تخفیف‌ها');
    await expect(root(page).locator('[data-shop-sub="discounts"] [data-shop-tool]')).toHaveText(['پیش‌نمایش بازیکن', 'تنظیمات']);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input:visible').fill('تخفیف برچسب');
    await expect(page.locator('.pl-pal-row').first()).toContainText('تخفیف برچسب‌دارها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/shop\/discounts/);
  });

  test('a prize gives a player tag for 30 days', async ({ page }) => {
    await open(page, '#/economy/prizes/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(k + '_vip30');
    await fillTranslated(p, 'name', { fa: 'ویژهٔ ۳۰ روزه ' + k.slice(-4) });
    const row = p.locator('.eco-items-row').first();
    await row.locator('select').first().selectOption('tag');
    await pickIn(row.locator('[data-tag-picker]'), 'ویژه ' + k.slice(-4));
    await expect(row.locator('[data-tag-days-hint]')).toContainText('۰ یعنی بی‌پایان');
    await expect(row.locator('select').nth(1).locator('option')).toHaveText(['اضافه به زمان باقی‌مانده', 'شروع دوباره از الان']);
    await savePanel(page);
    const prize = await byKey('prizes', k + '_vip30');
    made.prizes.push(prize.id);
    expect(prize.items).toEqual([{ type: 'tag', key: k + '_vip', action: 'give', value: 30 }]);
    await search(page, k + '_vip30');
    await expect(rows(page).first()).toContainText('برچسب زمان‌دار: ' + k + '_vip · ۳۰ روز');
  });

  test('a 5% discount for the tag is made in the editor and the player preview shows the new price', async ({ page }) => {
    await open(page, '#/economy/tag-discounts/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(k + '_vip5');
    await fillTranslated(p, 'name', { fa: '۵٪ ویژه ' + k.slice(-4) });
    await field(p, 'درصد تخفیف').locator('input').fill('۵');
    await pickRef(p, 'برچسب بازیکن', 'ویژه ' + k.slice(-4));
    await pickRef(p, 'آیتم‌های انتخابی', 'بستهٔ استیکر ' + k.slice(-4));
    await savePanel(page);
    const rule = await byKey('tag-discounts', k + '_vip5');
    expect(rule).toMatchObject({ percent: 5, tagIds: [vip.id], itemKeys: [item.key], allItems: false });
    await search(page, k + '_vip5');
    await expect(rows(page).first().locator('[data-td-percent]')).toHaveText('۵٪');
    await expect(rows(page).first().locator('[data-td-targets]')).toHaveText('۱ آیتم یا نوع آیتم');

    const u = sql(`INSERT INTO users (id, username) VALUES (gen_random_uuid(), '${k}_p') RETURNING id::text`).split('\n')[0].trim();
    try {
      sql(`INSERT INTO base_user_tag (user_id, tag_id, expires_at) VALUES ('${u}', '${vip.id}', now() + interval '30 days')`);
      await open(page, '#/economy/tag-discounts/preview?userId=' + u);
      const box = root(page).locator('[data-td-preview]');
      await expect(box.locator('[data-td-rules="1"]')).toContainText('۵٪');
      await expect(box.locator('[data-td-item="' + item.key + '"]')).toBeVisible();
      await expect(box).toContainText('۱۹۰');
    } finally {
      sql(`DELETE FROM users WHERE id='${u}'`);
    }
  });

  test('discount settings save the combine rule and the cap', async ({ page }) => {
    await open(page, '#/economy/tag-discounts/settings');
    const form = page.locator('[data-settings-form="tag-discounts"]');
    await field(form, 'چند تخفیف هم‌زمان').locator('select').selectOption('sum');
    await field(form, 'سقف تخفیف (درصد)').locator('input').fill('۳۰');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    expect((await eco.get('/tag-discount-settings')).settings).toMatchObject({ enabled: true, combine: 'sum', maxPercent: 30, withOwnDiscount: 'stack' });
  });

  test('a chest has its kind, daily and total limits and hidden odds', async ({ page }) => {
    const prize = await eco.post('/prizes', { key: k + '_c5', name: { fa: '۵ سکه' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 5 }] });
    made.prizes.push(prize.id);
    await open(page, '#/economy/chance-games/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(k + '_chest');
    await fillTranslated(p, 'name', { fa: 'صندوقچه ' + k.slice(-4) });
    await field(p, 'شکل بازی').locator('select').selectOption('chest');
    await field(p, 'سقف هر روز برای هر بازیکن').locator('input').fill('۳');
    await field(p, 'سقف کل برای هر بازیکن').locator('input').fill('۱۰');
    await p.locator('label.x-toggle', { hasText: 'شانس هر جایزه به بازیکن نشان داده شود' }).click();
    await p.locator('[data-col="prizeId"]').first().locator('button[aria-haspopup="listbox"]').first().click();
    const pop = p.locator('[data-col="prizeId"]').first().locator('.eco-pop').first();
    await pop.locator('input').fill(k + '_c5');
    await pop.locator('[role="option"]').first().click();
    await savePanel(page);
    const g = await byKey('chance-games', k + '_chest');
    made.games.push(g.id);
    expect(g).toMatchObject({ kind: 'chest', dailyLimit: 3, totalLimit: 10, showOdds: false });
    await search(page, k + '_chest');
    await expect(rows(page).first().locator('[data-ch-kind="chest"]')).toContainText('صندوقچه');
    await expect(rows(page).first().locator('[data-ch-daily]')).toContainText('روزی ۳ بار');
  });

  test('the new pages fit a phone screen and read in English', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const hash of ['#/economy/tag-discounts', '#/economy/tag-discounts/new', '#/economy/tag-discounts/preview', '#/economy/tag-discounts/settings', '#/economy/chance-games/new']) {
      await open(page, hash);
      await settle(page);
      await test.step(hash, () => noHorizontalScroll(page));
    }
    await page.setViewportSize({ width: 1280, height: 900 });
    await open(page, '#/economy/tag-discounts');
    await page.locator('.pl-tb-end .pl-seg-lang [data-locale="en"]').click();
    await expect(root(page).locator('[data-shop-tab="discounts"]')).toHaveText('Discounts');
    await expect(root(page).locator('[data-shop-sub="discounts"] [data-shop-tool]')).toHaveText(['Player preview', 'Settings']);
    await page.locator('.pl-tb-end .pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });
});
