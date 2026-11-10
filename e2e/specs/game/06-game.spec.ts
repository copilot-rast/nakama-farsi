import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, pub, rpc, uid } from '../../lib/api';
import { API_URL, GAME_URL } from '../../lib/env';
import { gameRpc, openGame, persian, screen, toast, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, opts: { query?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-g-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

async function steer(page: Page) {
  await page.evaluate(() => {
    const w = window as any;
    clearInterval(w.__e2eSteer);
    w.__e2eSteer = setInterval(() => {
      const g = w.__stargame;
      if (g.phase !== 'playing') return;
      const b = g.basket;
      let best: any = null;
      let rank = -1e9;
      for (const o of g.objects) {
        if (o.kind === 'bomb' || o.y > b.y + 4) continue;
        const r = o.y + (o.kind === 'gold' ? 120 : 0) - Math.abs(o.x - b.x) * 0.35;
        if (r > rank) { rank = r; best = o; }
      }
      let x = best ? best.x : b.x;
      for (const o of g.objects) {
        if (o.kind === 'bomb' && o.y > b.y - 110 && o.y < b.y + 12 && Math.abs(o.x - x) < b.w / 2 + 22) x = o.x < g.width / 2 ? o.x + b.w : o.x - b.w;
      }
      const c = document.getElementById('canvas')!;
      const rect = c.getBoundingClientRect();
      c.dispatchEvent(new PointerEvent('pointermove', { clientX: rect.left + x, clientY: rect.top + 10, pointerType: 'mouse', bubbles: true }));
    }, 25);
  });
}

async function openShopTab(page: Page, cat: string) {
  await page.locator('#screen-home [data-go="shop"]').click();
  await screen(page, 'shop');
  await page.locator(`#shop-tabs [data-cat="${cat}"]`).click();
  await expect(page.locator(`#shop-tabs [data-cat="${cat}"]`)).toHaveAttribute('aria-selected', 'true');
}

async function backHome(page: Page, from: string) {
  await page.locator(`#screen-${from} [data-go="home"]`).first().click();
  await screen(page, 'home');
}

test.describe('stargame', () => {
  test('a new device gets 50 coins, a full round pays coins, a replayed finish is refused', async ({ page }) => {
    const me = await start(page, 'round');
    await expect(page.locator('#balance')).toHaveText('۵۰');
    await expect(page.locator('#username')).toHaveText(me.username);
    await page.locator('#btn-play').click();
    await screen(page, 'play');
    await page.waitForFunction(() => (window as any).__stargame.phase === 'playing', null, { timeout: 15_000 });
    await steer(page);
    await expect(page.locator('#hud-time')).not.toHaveText('۳۰');
    await page.waitForFunction(() => (window as any).__stargame.phase === 'done', null, { timeout: 60_000 });
    const overlay = page.locator('#stage-overlay');
    await expect(overlay).toContainText('دور تمام شد!');
    const coins = Number(await overlay.getAttribute('data-coins'));
    const score = await page.evaluate(() => (window as any).__stargame.score);
    test.info().annotations.push({ type: 'round', description: `score ${score}, coins ${coins}` });
    expect(score, 'steering should catch some stars').toBeGreaterThanOrEqual(5);
    expect(coins).toBe(Math.min(30, Math.floor(score / 5)));
    await expect(overlay).toContainText(`${persian(coins)} سکه گرفتی. موجودی: ${persian(50 + coins)} سکه`);
    await expect(page.locator('#balance')).toHaveText(persian(50 + coins));
    const wallet = await rpc(me, 'economy_balances', {});
    expect(wallet.data.wallet.coin).toBe(50 + coins);
    const roundId = await page.evaluate(() => (window as any).eval('game.roundId'));
    const again = await gameRpc(page, 'stargame_finish', { roundId, score: 150 });
    expect(again).toMatchObject({ ok: false, code: 'ROUND_ALREADY_FINISHED' });
    const after = await rpc(me, 'economy_balances', {});
    expect(after.data.wallet.coin).toBe(50 + coins);
    await overlay.getByRole('button', { name: 'خانه' }).click();
    await screen(page, 'home');
  });

  test('STAR100 once, buy a skin with coins, then INSUFFICIENT_BALANCE', async ({ page }) => {
    await start(page, 'shop');
    await page.locator('#screen-home [data-go="gift"]').click();
    await screen(page, 'gift');
    await page.locator('#gift-code').fill('star100');
    await page.locator('#gift-form button[type="submit"]').click();
    await expect(page.locator('#gift-msg')).toHaveClass(/ok/);
    await expect(page.locator('#gift-msg')).toHaveText('هدیه گرفتی: ۱۰۰ سکه');
    await expect(page.locator('#balance')).toHaveText('۱۵۰');
    await page.locator('#gift-code').fill('STAR100');
    await page.locator('#gift-form button[type="submit"]').click();
    await expect(page.locator('#gift-msg')).toHaveClass(/bad/);
    await expect(page.locator('#gift-msg')).toHaveText('این کد را قبلاً استفاده کرده‌ای.');
    await expect(page.locator('#balance')).toHaveText('۱۵۰');
    await page.locator('#gift-code').fill('NOPE' + k.slice(-4).toUpperCase());
    await page.locator('#gift-form button[type="submit"]').click();
    await expect(page.locator('#gift-msg')).toHaveText('این کد وجود ندارد.');
    await backHome(page, 'gift');

    await openShopTab(page, 'star_skins');
    const gold = page.locator('#products [data-key="star_skin_gold"]');
    await expect(gold).toContainText('۸۰ سکه');
    await gold.locator('button[data-method="in_app"]').click();
    const dlg = page.locator('#dialog');
    await expect(dlg).toBeVisible();
    await expect(dlg).toContainText('۸۰ سکه');
    await dlg.getByRole('button', { name: 'بخر' }).click();
    await expect(toast(page)).toContainText('خرید موفق!');
    await expect(toast(page)).toContainText('سبد طلایی');
    await expect(page.locator('#balance')).toHaveText('۷۰');
    await expect(gold).toContainText('مال توست');

    const neon = page.locator('#products [data-key="star_skin_neon"]');
    await neon.locator('button[data-method="in_app"]').click();
    await dlg.getByRole('button', { name: 'بخر' }).click();
    await expect(toast(page)).toHaveClass(/bad/);
    await expect(toast(page)).toHaveText('سکه‌ات کافی نیست. چند دور بازی کن یا بسته‌ی سکه بخر.');
    await expect(page.locator('#balance')).toHaveText('۷۰');
    await expect(neon.locator('button[data-method="in_app"]')).toBeVisible();

    await backHome(page, 'shop');
    await expect(page.locator('#skins [data-skin="skin_gold"]')).toBeVisible();
    await page.locator('#screen-home [data-go="payments"]').click();
    await screen(page, 'payments');
    await expect(page.locator('#payments .pay[data-method="in_app"][data-status="success"]')).toHaveCount(1);
    await expect(page.locator('#payments .pay').first()).toContainText('۸۰ سکه');
  });

  test('bazaar and myket simulators pay; a used store token is refused', async ({ page }) => {
    const me = await start(page, 'store');
    await openShopTab(page, 'star_coin_packs');
    const pack = page.locator('#products [data-key="star_pack_small"]');
    await pack.locator('button[data-method="bazaar"]').click();
    const dlg = page.locator('#dialog');
    await expect(dlg.locator('.sim-badge')).toHaveText('شبیه‌ساز فروشگاه (حالت تست)');
    await expect(dlg).toContainText('star_pack_small');
    const bazaarToken = await dlg.locator('#sim-token').inputValue();
    expect(bazaarToken).toMatch(/^sandbox-[0-9a-f-]{36}-[0-9a-f]+$/);
    await dlg.locator('#sim-pay').click();
    await expect(dlg).toBeHidden();
    await expect(toast(page)).toContainText('خرید موفق! ۱۰۰ سکه');
    await expect(page.locator('#balance')).toHaveText('۱۵۰');

    await pack.locator('button[data-method="myket"]').click();
    await expect(dlg.locator('.sim-logo.myket')).toBeVisible();
    await dlg.locator('#sim-token').fill(bazaarToken);
    await dlg.locator('#sim-pay').click();
    await expect(dlg.locator('.form-msg')).toHaveClass(/bad/);
    const refused = await dlg.locator('.form-msg').innerText();
    expect(refused).toBe('این توکن مال این پرداخت یا این حساب نیست.');
    await expect(page.locator('#balance')).toHaveText('۱۵۰');
    const myketToken = await dlg.locator('#sim-token').evaluate((el: HTMLInputElement) => el.defaultValue);
    await dlg.locator('#sim-token').fill(myketToken);
    await dlg.locator('#sim-pay').click();
    await expect(dlg).toBeHidden();
    await expect(toast(page)).toContainText('خرید موفق! ۱۰۰ سکه');
    await expect(page.locator('#balance')).toHaveText('۲۵۰');

    const reuse = await gameRpc(page, 'economy_purchase_review', { method: 'bazaar', sku: 'star_pack_small', token: bazaarToken });
    expect(reuse).toMatchObject({ ok: true, data: { status: 'success', rewardGranted: true } });
    expect((await rpc(me, 'economy_balances', {})).data.wallet.coin).toBe(250);
    const start2 = await gameRpc(page, 'economy_purchase_start', { productId: (await rpc(me, 'economy_catalog', {})).data.products.find((p: any) => p.key === 'star_pack_small').id, method: 'bazaar' });
    const dup = await gameRpc(page, 'economy_purchase_submit', { paymentId: start2.data.paymentId, token: bazaarToken });
    expect(dup).toMatchObject({ ok: false, code: 'DEVELOPER_PAYLOAD_MISMATCH' });
    await page.locator('#btn-restore').click();
    await expect(dlg.locator('#restore-token')).toHaveValue(myketToken);
    await expect(dlg.locator('#restore-method')).toHaveValue('myket');
    await dlg.getByRole('button', { name: 'بررسی و بازیابی' }).click();
    await expect(dlg.locator('.form-msg')).toHaveClass(/ok/);
    await expect(dlg.locator('.form-msg')).toContainText('دوباره داده نمی‌شود');
    await dlg.getByRole('button', { name: 'بستن' }).click();
    await expect(page.locator('#balance')).toHaveText('۲۵۰');
  });

  test('HALF50 halves a zibal start: the sandbox gateway asks for half', async ({ page }) => {
    await start(page, 'zibal');
    await openShopTab(page, 'star_coin_packs');
    await expect(page.locator('#discount-row')).toBeVisible();
    await page.locator('#discount-code').fill('half50');
    const pack = page.locator('#products [data-key="star_pack_small"]');
    await expect(pack).toContainText('۱۰٬۰۰۰ تومان');
    const nav = page.waitForURL(/gateway\.zibal\.ir\/start\//, { timeout: 30_000 }).then(() => true, () => false);
    await pack.locator('button[data-method="zibal"]').click();
    test.skip(!(await nav), 'Zibal sandbox gateway (gateway.zibal.ir) is not reachable from this Mac');
    await expect(page.locator('body')).toContainText('50,000', { timeout: 20_000 });
    await expect(page.locator('body')).toContainText('ریال');
  });

  test('zarinpal sandbox: HALF50 amount, «پرداخت» returns to the game with success and coins', async ({ page }) => {
    const me = await start(page, 'zarin');
    const game = new URL(GAME_URL);
    const api = new URL(API_URL);
    await page.route((u) => u.searchParams.has('payment') && u.hostname === api.hostname && u.origin !== game.origin, (route) => {
      const u = new URL(route.request().url());
      return route.fulfill({ status: 302, headers: { location: GAME_URL + u.search } });
    });
    await openShopTab(page, 'star_coin_packs');
    await page.locator('#discount-code').fill('HALF50');
    const reached = page.waitForURL(/sandbox\.zarinpal\.com\/pg\/StartPay\//, { timeout: 30_000 }).then(() => true, () => false);
    await page.locator('#products [data-key="star_pack_small"] button[data-method="zarinpal"]').click();
    test.skip(!(await reached), 'ZarinPal sandbox (sandbox.zarinpal.com) is not reachable from this Mac');
    await expect(page.locator('body')).toContainText('50,000 ریال', { timeout: 20_000 });
    await page.locator('button', { hasText: /^پرداخت$/ }).first().click();
    await page.waitForURL((u) => u.href.startsWith(GAME_URL), { timeout: 45_000 });
    const result = page.locator('#dialog-body[data-payment-result]');
    await expect(result).toHaveAttribute('data-payment-result', 'success', { timeout: 60_000 });
    await expect(result).toContainText('پرداخت موفق بود');
    await expect(result).toContainText('موجودی: ۱۵۰ سکه');
    await expect(page.locator('#balance')).toHaveText('۱۵۰');
    await result.getByRole('button', { name: 'باشه' }).click();
    await page.locator('#screen-home [data-go="payments"]').click();
    const row = page.locator('#payments .pay[data-method="zarinpal"]').first();
    await expect(row).toHaveAttribute('data-status', 'success');
    await expect(row).toContainText('۵٬۰۰۰ تومان');
    await expect(row).toContainText('کد پیگیری');
    const mine = await rpc(me, 'economy_payments_mine', {});
    expect(mine.data.items[0]).toMatchObject({ method: 'zarinpal', status: 'success', finalToman: 5000 });
  });

  test('language switch fa → en → ar changes texts, direction and server content', async ({ page }) => {
    const me = await start(page, 'lang');
    const html = page.locator('html');
    await expect(html).toHaveAttribute('dir', 'rtl');
    await expect(page.locator('#btn-play')).toHaveText('شروع بازی');
    await expect(page.locator('#lang-picker')).toBeVisible();
    await page.locator('#lang-options [data-lang="en"]').click();
    await expect(html).toHaveAttribute('lang', 'en');
    await expect(html).toHaveAttribute('dir', 'ltr');
    await expect(page.locator('#btn-play')).toHaveText('Play');
    await expect(page.locator('#balance')).toHaveText('50');
    await expect(page.locator('#lang-options [data-lang="en"]')).toHaveAttribute('aria-pressed', 'true');
    await openShopTab(page, 'star_coin_packs');
    const enCat = await rpc(me, 'economy_catalog', { lang: 'en' });
    const enName = enCat.data.products.find((p: any) => p.key === 'star_pack_small').name;
    await expect(page.locator('#products [data-key="star_pack_small"] h3')).toHaveText(enName);
    await backHome(page, 'shop');
    await page.locator('#lang-options [data-lang="ar"]').click();
    await expect(html).toHaveAttribute('lang', 'ar');
    await expect(html).toHaveAttribute('dir', 'rtl');
    await expect(page.locator('html')).toHaveClass(/ui-fallback/);
    await expect(page.locator('#btn-play')).toHaveText('Play');
    await openShopTab(page, 'star_coin_packs');
    const arCat = await rpc(me, 'economy_catalog', { lang: 'ar' });
    const arName = arCat.data.products.find((p: any) => p.key === 'star_pack_small').name;
    expect(arName).not.toBe(enName);
    await expect(page.locator('#products [data-key="star_pack_small"] h3')).toHaveText(arName);
    await page.reload();
    await expect(page.locator('#screen-home')).toBeVisible();
    await expect(html).toHaveAttribute('lang', 'ar');
    await page.locator('#lang-options [data-lang="fa"]').click();
    await expect(html).toHaveAttribute('lang', 'fa');
    await expect(page.locator('#btn-play')).toHaveText('شروع بازی');
  });

  test('news, faq and inbox screens; an inbox action opens the shop', async ({ page }) => {
    const me = await start(page, 'screens');
    await page.locator('#menu-news').click();
    await screen(page, 'news');
    await expect(page.locator('#news-list .skel')).toHaveCount(0);
    const cards = page.locator('#news-list .news-card');
    if (await cards.count()) {
      await cards.first().click();
      await screen(page, 'article');
      await expect(page.locator('#article h1')).not.toBeEmpty();
      await page.locator('#screen-article [data-go="news"]').click();
      await screen(page, 'news');
    } else {
      await expect(page.locator('#news-list .empty')).toHaveText('هنوز خبری منتشر نشده.');
    }
    await backHome(page, 'news');
    await page.locator('#screen-home [data-go="faq"]').click();
    await screen(page, 'faq');
    await page.locator('#faq-q').fill('zz-no-such-question-' + k);
    await expect(page.locator('#faq-list .empty')).toHaveText('با این جستجو سؤالی پیدا نشد.');
    await backHome(page, 'faq');
    await page.locator('#btn-inbox').click();
    await screen(page, 'inbox');
    await expect(page.locator('#inbox-list .empty')).toHaveText('صندوقت خالی است.');
    await expect(page.locator('#btn-seen-all')).toBeDisabled();

    await pub.post('/notifications', { channel: 'in_app', title: { fa: 'هدیهٔ فروشگاه ' + k }, body: { fa: 'یک سر به **فروشگاه** بزن.' }, action: { type: 'goto', value: 'shop' }, target: { mode: 'users', userIds: [me.id] } });
    await expect(page.locator('#note-toast')).toContainText('هدیهٔ فروشگاه ' + k, { timeout: 20_000 });
    const item = page.locator('#inbox-list .inbox-item', { hasText: 'هدیهٔ فروشگاه ' + k });
    await expect(item).toHaveClass(/unread/);
    await expect(page.locator('#inbox-badge')).toHaveText('۱');
    await item.locator('[data-action="goto"]').click();
    await screen(page, 'shop');
    await expect(page.locator('#inbox-badge')).toBeHidden();
  });
});

