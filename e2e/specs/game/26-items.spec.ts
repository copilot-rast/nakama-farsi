import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, eco, uid } from '../../lib/api';
import { noHorizontalScroll } from '../../lib/console';
import { balance, gameRpc, openGame, persian, screen, toast, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-it-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

async function openInventory(page: Page) {
  await page.evaluate(() => (window as any).eval('show("inventory")'));
  await screen(page, 'inventory');
  await expect(page.locator('#inventory-list .spinner')).toHaveCount(0);
  await expect(page.locator('[data-item="star_basket_classic"]')).toBeVisible();
}

async function openShopTab(page: Page, cat: string) {
  await page.evaluate(() => (window as any).eval('show("shop")'));
  await screen(page, 'shop');
  await page.locator(`#shop-tabs [data-cat="${cat}"]`).click();
  await expect(page.locator('#products .spinner')).toHaveCount(0);
}

test.describe('stargame: inventory and product rules', () => {
  test('buy a skin in the shop, it is owned, equip it, the console player card shows it; the one-time pack leaves the shop', async ({ page, spawn }) => {
    test.setTimeout(120_000);
    const me = await start(page, 'buy');
    const gift = await gameRpc(page, 'economy_gift_redeem', { code: 'STAR100' });
    expect(gift.ok).toBe(true);
    await page.evaluate(() => (window as any).eval('refreshBalances()'));
    await expect(page.locator('#balance')).toHaveText(persian(150));

    await openInventory(page);
    const classic = page.locator('[data-item="star_basket_classic"]');
    await expect(classic.locator('[data-inv-equipped]')).toHaveText('در بازی');
    const ruby = page.locator('[data-item="star_basket_ruby"]');
    await expect(ruby).not.toHaveAttribute('data-owned', '');
    await expect(ruby.locator('[data-inv-note]')).toHaveText('در فروشگاه');
    await expect(page.locator('[data-item="star_frame_gold"] [data-inv-note]')).toHaveText('از سطح ۲');
    await expect(page.locator('[data-item="star_basket_ice"] [data-inv-buy]')).toHaveText('خرید با ۱۲۰ سکه');
    await expect(page.locator('#inventory-count')).toHaveText('۱ از ۴');

    await openShopTab(page, 'star_skins');
    const card = page.locator('#products [data-key="star_item_ruby"]');
    await card.locator('[data-method="in_app"]').click();
    await page.locator('#dialog .btn-primary').click();
    await expect(toast(page)).toContainText('سبد یاقوتی مال تو شد');
    await expect(card.locator('[data-owned]')).toHaveText('مال توست');
    await expect(page.locator('#balance')).toHaveText(persian(90));
    const again = await gameRpc(page, 'economy_purchase_start', { productId: (await eco.get('/products?q=star_item_ruby')).items[0].id, method: 'in_app' });
    expect(again).toMatchObject({ ok: false, code: 'PRIZE_ALREADY_OWNED' });

    await openInventory(page);
    await expect(ruby).toHaveAttribute('data-owned', '');
    await ruby.locator('[data-inv-equip]').click();
    await expect(toast(page)).toContainText('سبد یاقوتی انتخاب شد');
    await expect(ruby.locator('[data-inv-equipped]')).toBeVisible();
    await expect(classic.locator('[data-inv-equip]')).toBeVisible();
    expect(await page.evaluate(() => localStorage.getItem('stargame.skin'))).toBe('skin_ruby');

    await openShopTab(page, 'star_coin_packs');
    const pack = page.locator('#products [data-key="star_starter_pack"]');
    await expect(pack.locator('[data-one-time]')).toHaveText('یک بار');
    await pack.locator('[data-method="in_app"]').click();
    await page.locator('#dialog .btn-primary').click();
    await expect(toast(page)).toContainText('خوش آمدی! ۱۵۰ سکه به کیف پولت اضافه شد.');
    await expect(pack).toHaveCount(0);
    expect(await balance(page)).toBe(210);
    const mine = await gameRpc(page, 'economy_purchases_mine', {});
    const keys = ((mine as any).data.items as any[]).map((x) => x.key).sort();
    expect(keys).toEqual(['star_item_ruby', 'star_starter_pack']);

    const c = await spawn({ signedIn: true });
    await c.goto(CONSOLE_URL + '/#/players/' + me.id);
    const inv = c.locator('[data-player-items]');
    await expect(inv).toBeVisible({ timeout: 20_000 });
    await expect(inv).toContainText('انبار بازیکن');
    await expect(inv.locator('[data-inv-item="star_basket_ruby"]')).toHaveAttribute('data-equipped', '');
    await expect(inv.locator('[data-inv-item="star_basket_ruby"]')).toContainText('جایزه');
    await expect(inv.locator('[data-inv-equipped]')).toContainText('سبد یاقوتی');
    await expect(inv.locator('[data-inv-counts]')).toHaveText('۲ از ۴ آیتم');
  });

  test('buy an item straight from the inventory, a second buy is refused, the round uses the equipped basket', async ({ page }) => {
    await start(page, 'direct');
    await gameRpc(page, 'economy_gift_redeem', { code: 'STAR100' });
    await openInventory(page);
    const ice = page.locator('[data-item="star_basket_ice"]');
    await ice.locator('[data-inv-buy]').click();
    await page.locator('#dialog .btn-primary').click();
    await expect(toast(page)).toContainText('سبد یخی مال تو شد');
    await expect(ice).toHaveAttribute('data-owned', '');
    await expect(page.locator('#balance')).toHaveText(persian(30));
    const dup = await gameRpc(page, 'economy_item_buy', { key: 'star_basket_ice' });
    expect(dup).toMatchObject({ ok: false, code: 'ALREADY_OWNED' });
    const locked = await gameRpc(page, 'economy_item_unlock', { key: 'star_frame_gold' });
    expect(locked).toMatchObject({ ok: false, code: 'LEVEL_TOO_LOW' });
    await ice.locator('[data-inv-equip]').click();
    await expect(ice.locator('[data-inv-equipped]')).toBeVisible();
    await page.locator('#screen-inventory [data-go="home"]').click();
    await screen(page, 'home');
    expect(await page.evaluate(() => (window as any).eval('selectedSkin()'))).toBe('skin_ice');
  });

  test('English inventory with Latin digits fits a phone screen', async ({ page }) => {
    await start(page, 'en', { lang: 'en' });
    await expect(page.locator('#menu-inventory')).toContainText('Inventory');
    await page.locator('#menu-inventory').click();
    await screen(page, 'inventory');
    await openInventory(page);
    await expect(page.locator('[data-item="star_basket_classic"]')).toContainText('Wooden basket');
    await expect(page.locator('[data-item="star_frame_gold"] [data-inv-note]')).toHaveText('From level 2');
    await expect(page.locator('#inventory-count')).toHaveText('1 of 4');
    await page.setViewportSize({ width: 390, height: 844 });
    await noHorizontalScroll(page);
  });
});
