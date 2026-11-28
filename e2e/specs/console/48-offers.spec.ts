import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { users: [] as string[], products: [] as string[], prizes: [] as string[], tags: [] as string[] };
let savedRow = '';
let tag: any;
let prod: any;
let P: Player;

test.describe.configure({ mode: 'serial' });

function restoreOfferSettings(row: string) {
  if (!row) {
    sql('DELETE FROM economy_offer_settings');
    return;
  }
  const i = row.lastIndexOf('|');
  const q = (v: string) => "'" + v.replace(/'/g, "''") + "'";
  sql('UPDATE economy_offer_settings SET settings=' + q(row.slice(0, i)) + '::jsonb, updated_at=' + q(row.slice(i + 1)) + '::timestamptz');
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
  savedRow = sql(`SELECT settings::text || '|' || updated_at::text FROM economy_offer_settings`);
  tag = await base.post('/tags', { key: 'e2e_' + k + '_oft', keyType: 'user', name: { fa: 'پیشنهادی ' + k.slice(-4) }, description: {}, isPublic: false, metadata: {}, sort: 0 });
  made.tags.push(tag.id);
  const prize = await eco.post('/prizes', { key: 'e2e_' + k + '_opz', name: { fa: 'جایزهٔ پیشنهاد' }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  prod = await eco.post('/products', { key: 'e2e_' + k + '_op', isActive: true, name: { fa: 'بستهٔ پیشنهادی ' + k, en: 'Offer pack ' + k }, prizeId: prize.id, paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: 200 } });
  made.products.push(prod.id);
  P = await devicePlayer('e2e-ofc-' + k, ('ofc' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await base.put('/players/' + P.id + '/tags', { tagIds: [tag.id] });
});

test.afterAll(async () => {
  restoreOfferSettings(savedRow);
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_offer WHERE key LIKE 'e2e_${k}%'`);
  const prods = made.products.map((x) => `'${x}'`).join(',');
  if (prods) sql(`DELETE FROM economy_payment_log WHERE payment_id IN (SELECT id FROM economy_payment WHERE product_id IN (${prods})); DELETE FROM economy_payment WHERE product_id IN (${prods}); DELETE FROM economy_user_prize WHERE source_type='payment' AND prize_id IN (SELECT prize_id FROM economy_product WHERE id IN (${prods}))`);
  for (const id of made.products) await eco.del('/products/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.tags) await base.del('/tags/' + id);
});

test.describe('offers: hub, editors and settings', () => {
  test('the shop has the offers tab; a daily offer is built in its editor and listed', async ({ page }) => {
    await open(page, '#/economy/shop/products');
    await expect(root(page).locator('[data-shop-tab="offers"]')).toHaveText('پیشنهادها');
    await open(page, '#/economy/offers');
    const sub = root(page).locator('[data-shop-sub="offers"]');
    await expect(sub.locator('[data-shop-list]')).toHaveText(['روزانه', 'پیشنهاد ورود']);
    await expect(sub.locator('[data-shop-tool]')).toHaveText(['پیش‌نمایش بازیکن', 'گزارش', 'تنظیمات']);
    await root(page).getByRole('button', { name: 'پیشنهاد روزانهٔ تازه' }).first().click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill('e2e_' + k + '_daily');
    await fillTranslated(p, 'name', { fa: 'روزانهٔ ' + k, en: 'Daily ' + k });
    await pickIn(field(p, /^محصول$/), 'e2e_' + k + '_op');
    await field(p, 'درصد تخفیف').locator('input').fill('۳۰');
    await field(p, /^سقف خرید$/).locator('input').fill('۲');
    await field(p, /^وزن$/).locator('input').fill('۱۰۰۰۰۰۰');
    await pickIn(field(p, 'فقط برای این بازیکنان'), 'e2e_' + k + '_oft');
    await savePanel(page);
    const saved = await byKey('offers', 'e2e_' + k + '_daily');
    expect(saved).toMatchObject({ kind: 'daily', productId: prod.id, priceMode: 'discount', discountPercent: 30, purchaseLimit: 2, weight: 1000000, audienceTagIds: [tag.id], isActive: true });
    await search(page, 'e2e_' + k + '_daily');
    const row = rows(page).first();
    await expect(row).toContainText('روزانهٔ ' + k);
    await expect(row.locator('[data-offer-price="discount"]')).toContainText('۳۰٪ تخفیف');
    await expect(row).toContainText('ویژهٔ برچسب');
  });

  test('an entry offer with a lapsed-return trigger shows its trigger and timer as it is set', async ({ page }) => {
    await open(page, '#/economy/offers/entry/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill('e2e_' + k + '_back');
    await fillTranslated(p, 'name', { fa: 'خوش برگشتی ' + k });
    await pickIn(field(p, /^محصول$/), 'e2e_' + k + '_op');
    await field(p, /^قیمت$/).locator('select').selectOption('price');
    await pickIn(field(p, 'ارز'), 'coin');
    await field(p, 'مقدار').first().locator('input').fill('۵۰');
    await field(p, 'شرط شروع').locator('select').selectOption('lapsed');
    await field(p, 'عدد شرط').locator('input').fill('۷');
    await field(p, 'مدت (دقیقه)').locator('input').fill('۱۵۰');
    await expect(p.locator('[data-offer-trigger-say="lapsed"]')).toContainText('بعد از دست‌کم ۷ روز غیبت');
    await expect(p.locator('[data-offer-trigger-say="lapsed"]')).toContainText('۲ ساعت ۳۰ دقیقه');
    await field(p, 'دوباره باز شود').locator('select').selectOption('cooldown');
    await field(p, 'وقفه (ساعت)').locator('input').fill('۴۸');
    await savePanel(page);
    const saved = await byKey('entry-offers', 'e2e_' + k + '_back');
    expect(saved).toMatchObject({ kind: 'entry', triggerType: 'lapsed', triggerValue: 7, durationMinutes: 150, repeat: 'cooldown', cooldownHours: 48, priceMode: 'price', coinPrice: { currencyKey: 'coin', amount: 50 } });
    await search(page, 'e2e_' + k + '_back');
    await expect(rows(page).first()).toContainText('بعد از دست‌کم ۷ روز غیبت');
    await expect(rows(page).first()).toContainText('۲ ساعت ۳۰ دقیقه');
  });

  test('settings save the period, count, reroll cost and timer mode, and refuse a bad value', async ({ page }) => {
    await open(page, '#/economy/offers/settings');
    const form = page.locator('[data-settings-form="offers"]');
    await expect(form.locator('[data-offer-period]')).toContainText('دورهٔ فعلی');
    await field(form, /^ساعت شروع روز$/).locator('input').fill('۳۰');
    await form.locator('[data-settings-save]').click();
    await expect(field(form, /^ساعت شروع روز$/)).toContainText('عدد');
    await field(form, /^ساعت شروع روز$/).locator('input').fill('۴');
    await field(form, /^تعداد پیشنهاد$/).locator('input').fill('۱');
    await pickIn(field(form, 'ارز هزینه'), 'coin');
    await field(form, /^هزینه$/).locator('input').fill('۵');
    await field(form, 'شروع زمان‌سنج').locator('select').selectOption('seen');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    const st = (await eco.get('/offer-settings')).settings;
    expect(st).toMatchObject({ resetHour: 4, count: 1, rerollCost: { currencyKey: 'coin', amount: 5 }, entryTimerStart: 'seen', timezone: 'Asia/Tehran', weighting: 'weighted' });
  });
});

test.describe('offers: preview, player card and report', () => {
  test('the preview shows what the player gets with a reason for every offer', async ({ page }) => {
    await eco.put('/offer-settings', { ...(await eco.get('/offer-settings')).settings, count: 1, weighting: 'priority', entryTimerStart: 'list' });
    await open(page, '#/economy/offers/preview?userId=' + P.id);
    const box = root(page).locator('[data-offer-preview]');
    await expect(box.locator('[data-offer-summary]')).toContainText(P.username);
    await expect(box.locator(`[data-offer-draw="e2e_${k}_daily"]`)).toBeVisible();
    const row = box.locator('tr', { has: page.locator(`[data-offer-draw="e2e_${k}_daily"]`) });
    await expect(row.locator('[data-offer-reason]')).toHaveAttribute('data-offer-reason', 'picked');
    await expect(box.locator(`[data-offer-entry="e2e_${k}_back"]`)).toBeVisible();
  });

  test('the preview fits a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await open(page, '#/economy/offers/preview?userId=' + P.id);
    const box = root(page).locator('[data-offer-preview]');
    await expect(box.locator(`[data-offer-draw="e2e_${k}_daily"]`)).toBeVisible();
    await noHorizontalScroll(page);
  });

  test('the player card lists today offers; close and draw again work; the report counts the purchase', async ({ page }) => {
    const out = await rpc(P, 'economy_offers', { lang: 'fa' });
    const mine = out.data.daily.items.find((x: any) => x.key === 'e2e_' + k + '_daily');
    expect(mine).toMatchObject({ discountPercent: 30, coinPrice: { amount: 140 }, purchaseLimit: 2 });
    sql(`UPDATE users SET wallet = jsonb_set(COALESCE(wallet,'{}'::jsonb), '{coin}', '1000') WHERE id='${P.id}'`);
    const bought = await rpc(P, 'economy_offer_buy', { offerId: mine.offerId, method: 'in_app' });
    expect(bought.data.status).toBe('success');

    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-offers]');
    await expect(card.locator('h2')).toBeVisible({ timeout: 20_000 });
    const line = card.locator(`[data-pp-offer="e2e_${k}_daily"]`);
    await expect(line).toBeVisible();
    await expect(line).toContainText('فعال');
    await expect(card.locator('[data-offer-counts]')).toContainText('خریده ۱');
    await card.locator(`[data-offer-close="e2e_${k}_daily"]`).click();
    await page.locator('[data-economy="dialog"] button', { hasText: 'بستن' }).click();
    await expect(toasts(page).last()).toContainText('پیشنهاد بسته شد');
    await expect(card.locator(`[data-pp-offer="e2e_${k}_daily"]`)).toHaveAttribute('data-offer-state', 'closed');
    await card.locator('[data-offer-redraw]').click();
    await page.locator('[data-economy="dialog"] button', { hasText: 'قرعهٔ دوباره' }).click();
    await expect(toasts(page).last()).toContainText('پیشنهادهای امروز بسته شد');
    expect(sql(`SELECT count(*) FROM economy_user_offer_period WHERE user_id='${P.id}'`)).toBe('0');
    await expect(card.locator('[data-offer-preview-link]')).toHaveAttribute('href', new RegExp('offers/preview\\?userId=' + P.id));

    await open(page, '#/economy/offers/report?days=7');
    const rep = root(page).locator(`[data-offer-report="e2e_${k}_daily"]`);
    await expect(rep).toBeVisible();
    const tr = root(page).locator('tr', { has: page.locator(`[data-offer-report="e2e_${k}_daily"]`) });
    await expect(tr).toContainText('۱');
    await expect(root(page).locator(`[data-offer-conv="e2e_${k}_daily"]`)).toContainText('۱۰۰');
    await expect(root(page).locator('[data-offer-stats]')).toContainText('نرخ تبدیل');
  });
});
