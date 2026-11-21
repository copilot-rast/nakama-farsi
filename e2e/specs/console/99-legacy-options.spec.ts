import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { base, deleteUser, devicePlayer, eco, pub, rpc, uid, type Player } from '../../lib/api';
import { restoreConfig, sql } from '../../lib/db';
import { field, fillTranslated, go, open, openSettings, panel, root, savePanel, settlePanel, toasts } from '../../lib/console';

const k = uid();
const chestKey = 'e2e_' + k + '_chest';
const wheelKey = 'e2e_' + k + '_wheel';
const made = { users: [] as string[], categories: [] as string[] };
let P: Player;

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

async function pickIn(scope: Locator, text: string, exact?: RegExp) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = exact ? pop.locator('[role="option"]').filter({ has: scope.page().locator('span.font-mono', { hasText: exact }) }).first() : pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-lo-' + k, ('lo' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  for (const id of made.categories) await base.del('/categories/' + id);
  sql(`DELETE FROM economy_reward_claim WHERE owner_id IN (SELECT id FROM economy_streak WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (SELECT id FROM economy_prize WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/streaks?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/streaks/' + it.id)).status).toBe(200);
  for (const it of (await eco.get('/entry-offers?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/entry-offers/' + it.id)).status).toBe(200);
  for (const it of (await eco.get('/products?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/products/' + it.id)).status).toBe(200);
  sql(`DELETE FROM economy_chance_play WHERE game_id IN (SELECT id FROM economy_chance_game WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/chance-games?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/chance-games/' + it.id)).status).toBe(200);
  for (const it of (await eco.get('/prizes?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/prizes/' + it.id)).status).toBe(200);
});

test.describe('legacy options in the console', () => {
  test('an item type key in camelCase can be saved again', async ({ page }) => {
    const key = 'e2eSkin' + k.replace(/_/g, '').slice(-8);
    const cat = await base.post('/categories', { key, keyType: 'item_type', name: { fa: 'پوسته ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    made.categories.push(cat.id);
    await open(page, '#/project-settings/categories/' + cat.id);
    await settlePanel(page);
    const p = panel(page);
    await expect(field(p, 'کلید').locator('input')).toHaveValue(key);
    await fillTranslated(p, 'name', { fa: 'پوستهٔ تازه ' + k });
    await savePanel(page);
    const back = await base.get('/categories/' + cat.id);
    expect(back.key).toBe(key);
    expect(back.name.fa).toBe('پوستهٔ تازه ' + k);
  });

  test('a chest opened by a prize and a daily wheel with free turn first, day start and level gate', async ({ page }) => {
    const coinId = (await byKey('currencies', 'coin')).id;
    await open(page, '#/economy/chance-games/new');
    await settlePanel(page);
    let p = panel(page);
    await field(p, 'کلید').locator('input').fill(chestKey);
    await fillTranslated(p, 'name', { fa: 'صندوق قدیمی ' + k });
    await field(p, 'هر خانه چه می‌دهد').locator('select').selectOption('slots');
    const slot = p.locator('[data-rows="slots"] [data-row="0"]');
    await slot.locator('[data-col="chance"] input').fill('70');
    await slot.locator('[data-col="kind"] select').selectOption('currency');
    await pickIn(slot.locator('[data-col="currencyId"]'), 'coin', /^coin$/);
    await slot.locator('[data-col="min"] input').fill('120');
    await slot.locator('[data-col="max"] input').fill('180');
    await field(p, 'سقف تلاش دوباره').locator('input').fill('10000');
    await field(p, /^نوبت رایگان$/).locator('select').selectOption('daily');
    await field(p, 'نوبت رایگان هر چند دقیقه').locator('input').fill('');
    await savePanel(page);
    const chest = await byKey('chance-games', chestKey);
    expect(chest).toMatchObject({ prizeType: 'slots', rerollCap: 10000, freeReset: 'daily', intervalTime: null });

    await go(page, '#/economy/prizes/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_road');
    await fillTranslated(p, 'name', { fa: 'جایزهٔ جاده ' + k });
    const row = p.locator('.eco-items-row').first();
    await row.locator('select').first().selectOption('chest');
    await pickIn(row.locator('[data-chest-picker]').locator('..'), chestKey, new RegExp('^' + chestKey + '$'));
    await expect(row.locator('[data-chest-value]')).toContainText('همان لحظهٔ گرفتن جایزه');
    await savePanel(page);
    const road = await byKey('prizes', 'e2e_' + k + '_road');
    expect(road.items).toEqual([{ type: 'chest', key: chestKey, action: 'open', value: 1 }]);
    await eco.put('/prizes/' + road.id, Object.assign({}, road, { isRepeatable: true }));
    const before = (await rpc(P, 'economy_balances', {})).data.wallet.coin || 0;
    await eco.post('/users/' + P.id + '/grant', { prizeId: road.id });
    const after = (await rpc(P, 'economy_balances', {})).data.wallet.coin || 0;
    expect(after - before).toBeGreaterThanOrEqual(120);
    expect(after - before).toBeLessThanOrEqual(180);

    await go(page, '#/economy/prizes/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_rand');
    await fillTranslated(p, 'name', { fa: 'سکهٔ تصادفی ' + k });
    const r2 = p.locator('.eco-items-row').first();
    await pickIn(r2, 'coin', /^coin$/);
    await r2.locator('select').nth(1).selectOption('random');
    const rv = r2.locator('[data-random-value]');
    await rv.locator('input').nth(0).fill('150');
    await rv.locator('input').nth(1).fill('180');
    await savePanel(page);
    expect((await byKey('prizes', 'e2e_' + k + '_rand')).items).toEqual([{ type: 'currency', key: 'coin', action: 'random', value: { min: 150, max: 180 } }]);

    const win = (await eco.post('/prizes', { key: 'e2e_' + k + '_win', name: { fa: 'برد ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }] })).id;
    await go(page, '#/economy/chance-games/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill(wheelKey);
    await fillTranslated(p, 'name', { fa: 'گردونهٔ روزانه ' + k });
    const wrows = p.locator('[data-rows="prize"] [data-row]');
    for (let i = (await wrows.count()) - 1; i > 0; i--) await wrows.nth(i).locator('.eco-items-remove button').click();
    await wrows.nth(0).locator('[data-col="chance"] input').fill('100');
    await pickIn(wrows.nth(0).locator('[data-col="prizeId"]'), 'برد ' + k);
    await field(p, /^نوبت رایگان$/).locator('select').selectOption('daily');
    await field(p, 'نوبت رایگان هر چند دقیقه').locator('input').fill('');
    await p.locator('label.x-toggle', { hasText: 'اول نوبت رایگان' }).click();
    await pickIn(field(p, 'ارز بلیت'), 'coin', /^coin$/);
    await field(p, 'قیمت بلیت').locator('input').fill('10');
    await field(p, 'منطقهٔ زمانی روز').locator('input').fill('UTC');
    await field(p, /^شروع روز/).locator('input').fill('90');
    await field(p, 'کمترین سطح').locator('input').fill('5');
    await savePanel(page);
    const wheel = await byKey('chance-games', wheelKey);
    expect(wheel).toMatchObject({ freeReset: 'daily', freeFirst: true, dayZone: 'UTC', dayStartMinute: 90, minLevel: 5, price: { currencyId: coinId, amount: 10 } });
    const view = (await rpc(P, 'economy_chance_game', { id: wheel.id })).data.game;
    expect(view).toMatchObject({ freeReset: 'daily', freeFirst: true, minLevel: 5, locked: true, canPlay: false });
    const locked = await rpc(P, 'economy_chance_play', { id: wheel.id });
    expect(JSON.parse(locked.data.message).code).toBe('CHANCE_LOCKED');
  });

  test('streak: online days, 01:30 day start, same-day claims, sign-up delay and infinity by claim', async ({ page }) => {
    const day = (await eco.post('/prizes', { key: 'e2e_' + k + '_day', name: { fa: 'روز ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }] })).id;
    const inf = (await eco.post('/prizes', { key: 'e2e_' + k + '_inf', name: { fa: 'بی‌پایان ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'random', value: { min: 150, max: 180 } }] })).id;
    await open(page, '#/economy/streaks');
    await go(page, '#/economy/streaks/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_daily');
    await fillTranslated(p, 'name', { fa: 'روزانهٔ قدیمی ' + k });
    await pickIn(field(p, 'رویداد شمارنده'), 'online', /^online$/);
    await field(p, 'منطقهٔ زمانی روز').locator('input').fill('UTC');
    await field(p, /^شروع روز/).locator('input').fill('90');
    await field(p, 'مهلت گرفتن جایزه').locator('select').selectOption('period');
    await field(p, /^تأخیر بعد از ثبت‌نام/).locator('input').fill('1');
    const rewards = p.locator('[data-rows="rewards"]');
    await rewards.locator('[data-rows-add="rewards"]').click();
    await pickIn(rewards.locator('[data-row="0"] [data-col="prizeId"]'), 'روز ' + k);
    await p.locator('label.x-toggle', { hasText: 'بعد از پایان چرخه، مرحلهٔ بی‌پایان' }).click();
    await field(p, 'جایزهٔ مرحلهٔ بی‌پایان').locator('select').selectOption('claim');
    await expect(field(p, /^ارز$/)).toBeHidden();
    await pickIn(field(p, 'جایزهٔ هر روز بی‌پایان'), 'بی‌پایان ' + k);
    await savePanel(page);
    const sk = await byKey('streaks', 'e2e_' + k + '_daily');
    expect(sk).toMatchObject({ dayZone: 'UTC', dayStartMinute: 90, claimWindow: 'period', signupDelayDays: 1, infinity: true, infinityMode: 'claim', infinityPrizeId: inf, infinityCurrency: '' });
    expect(sk.rewards[0]).toMatchObject({ day: 1, prizeId: day });
  });

  test('daily offers by score, coin floor and an entry offer after a finished match', async ({ page }) => {
    try {
      await open(page, '#/economy/offer-settings');
      const f = root(page).locator('[data-settings-form="offers"]');
      await field(f, 'روش انتخاب').locator('select').selectOption('scored');
      await expect(f.locator('[data-os-scored]')).toBeVisible();
      await expect(field(f, 'امتیاز توان خرید').locator('input')).toHaveValue('۲۰');
      await field(f, 'امتیاز کمیابی').locator('input').fill('common=10, uncommon=10, rare=13, epic=13');
      await field(f, 'کمترین قیمت سکه‌ای بعد از تخفیف').locator('input').fill('1');
      await f.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toBeVisible();
      const st = (await eco.get('/offer-settings')).settings;
      expect(st).toMatchObject({ weighting: 'scored', coinPriceFloor: 1, score: { affordable: 20, recency: 18, newness: 16, recencyDays: 7, newDays: 7, floor: 1, rarity: { common: 10, uncommon: 10, rare: 13, epic: 13 } } });
    } finally {
      restoreConfig();
    }
    const games = (await eco.get('/games?limit=100')).items || [];
    test.skip(!games.length, 'no game on this server to trigger on');
    const prize = (await eco.post('/prizes', { key: 'e2e_' + k + '_np', name: { fa: 'تازه‌وارد ' + k }, items: [{ type: 'currency', key: 'coin', action: 'add', value: 5 }] })).id;
    const prod = (await eco.post('/products', { key: 'e2e_' + k + '_np', name: { fa: 'بستهٔ تازه‌وارد ' + k }, isActive: false, prizeId: prize, paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: 100 } })).id;
    await open(page, '#/economy/entry-offers');
    await go(page, '#/economy/entry-offers/new?productId=' + prod);
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_np');
    await field(p, 'شرط شروع').locator('select').selectOption('game');
    await field(p, 'عدد شرط').locator('input').fill('1');
    await pickIn(field(p, /^بازی$/), games[0].key);
    await expect(p.locator('[data-offer-trigger-say="game"]')).toContainText('بازی را تمام کرده باشد');
    await savePanel(page);
    const o = await byKey('entry-offers', 'e2e_' + k + '_np');
    expect(o).toMatchObject({ triggerType: 'game', triggerValue: 1, triggerRef: games[0].key });
  });

  test('friend day start, bot auto-decline, OTP per account and payment lifetime', async ({ page }) => {
    const pay = (await eco.get('/payment-settings')).settings;
    try {
      await openSettings(page, 'friends');
      const fr = root(page).locator('[data-settings-form="friends"]');
      await expect(fr).toContainText('روز اجتماعی');
      await field(fr, 'منطقهٔ زمانی روز').locator('input').fill('UTC');
      await field(fr, /^شروع روز/).locator('input').fill('90');
      await field(fr, 'رد خودکار درخواست به ربات').locator('input').fill('5');
      await fr.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toBeVisible();
      await expect.poll(() => sql(`SELECT day_zone || '/' || day_start_minute || '/' || bot_decline_hours FROM base_friend_rules WHERE id=1`)).toBe('UTC/90/5');
      const lim = await rpc(P, 'base_friend_rules', {});
      expect(lim.status).toBe(200);
      expect(lim.data).toMatchObject({ timeZone: 'UTC', dayStartMinute: 90 });

      await openSettings(page, 'otp');
      const otp = root(page).locator('[data-settings-form="otp"]');
      await field(otp, 'کد در ساعت برای هر حساب').locator('input').fill('5');
      await otp.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toBeVisible();
      expect((await pub.get('/otp/settings')).userHourly).toBe(5);

      await open(page, '#/economy/payment-settings');
      const ps = root(page).locator('[data-settings-form="payments"]');
      await field(ps, 'مهلت پرداخت باز').locator('input').fill('1440');
      await ps.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toBeVisible();
      expect((await eco.get('/payment-settings')).settings.pendingMinutes).toBe(1440);
    } finally {
      await eco.put('/payment-settings', pay);
      restoreConfig();
    }
  });
});
