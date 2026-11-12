import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { base, deleteUser, pub, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { gameRpc, openGame, screen, waitHome, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];
const extra: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of extra) await pub.del('/banners/' + id);
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-bn-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

function carousel(page: Page) {
  return page.locator('#banner-carousel');
}

async function slideKeys(page: Page) {
  return carousel(page).locator('.bn-slide').evaluateAll((els) => els.map((e) => e.getAttribute('data-banner')));
}

async function showSlide(page: Page, key: string) {
  const keys = await slideKeys(page);
  await carousel(page).locator('.bn-dot').nth(keys.indexOf(key)).click();
  const slide = carousel(page).locator(`.bn-slide[data-banner="${key}"]`);
  await expect(slide).toBeVisible();
  return slide;
}

function seen(userId: string, col: 'viewed_at' | 'clicked_at') {
  return Number(sql(`SELECT count(*) FROM publication_banner_seen s JOIN publication_banner b ON b.id=s.banner_id WHERE s.user_id='${userId}' AND s.${col} IS NOT NULL AND b.key LIKE 'star\\_bn\\_%'`));
}

test.describe('banners: Star Catcher', () => {
  test('menu carousel: seeded banners for a new player, rotation controls, views counted once, actions open the right screens', async ({ page }) => {
    const p = await start(page, 'new', { lang: 'fa' });
    await expect(carousel(page)).toBeVisible();
    expect(await slideKeys(page)).toEqual(['star_bn_welcome', 'star_bn_galaxy', 'star_bn_missions', 'star_bn_fan']);
    const first = carousel(page).locator('.bn-slide').first();
    await expect(first).toBeVisible();
    await expect(first).toContainText('به ستاره‌چین خوش آمدی');
    await expect(first.locator('.bn-cta')).toHaveText('کد هدیه');
    await expect(first.locator('img')).toHaveAttribute('src', /welcome-fa\.svg$/);
    await expect(carousel(page).locator('.bn-dot')).toHaveCount(4);
    await expect(carousel(page).locator('.bn-dot').first()).toHaveAttribute('aria-current', 'true');
    await expect(carousel(page).locator('.bn-dot').nth(1)).toHaveAttribute('aria-label', 'بنر ۲ از ۴');
    await carousel(page).locator('.bn-next').click();
    await expect(carousel(page).locator('.bn-slide[data-banner="star_bn_galaxy"]')).toBeVisible();
    await expect(first).toBeHidden();
    await expect.poll(() => seen(p.id, 'viewed_at')).toBe(2);

    await page.reload();
    await waitHome(page);
    await expect(carousel(page).locator('.bn-slide').first()).toBeVisible();
    await page.waitForTimeout(1500);
    expect(seen(p.id, 'viewed_at')).toBe(2);
    expect(Number(sql(`SELECT count(*) FROM publication_banner_seen WHERE user_id='${p.id}'`))).toBe(2);

    const galaxy = await showSlide(page, 'star_bn_galaxy');
    await galaxy.click();
    await screen(page, 'shop');
    await expect(page.locator('#products [data-key="star_skin_galaxy"]')).toHaveClass(/bn-flash/);
    await expect(page.locator('.tab[aria-selected="true"]')).toHaveAttribute('data-cat', 'star_skins');
    await page.locator('#screen-shop [data-go="home"]').click();
    await screen(page, 'home');

    await (await showSlide(page, 'star_bn_missions')).click();
    await screen(page, 'tasks');
    await page.locator('#screen-tasks [data-go="home"]').click();
    await screen(page, 'home');
    await (await showSlide(page, 'star_bn_fan')).click();
    await screen(page, 'achievements');
    await expect(page.locator('[data-achievement="star_ach_supporter"]')).toHaveClass(/bn-flash/);
    await page.locator('#screen-achievements [data-go="home"]').click();
    await screen(page, 'home');
    await (await showSlide(page, 'star_bn_welcome')).click();
    await screen(page, 'gift');
    await page.locator('#screen-gift [data-go="home"]').click();
    await screen(page, 'home');
    await (await showSlide(page, 'star_bn_galaxy')).click();
    await screen(page, 'shop');
    await expect.poll(() => seen(p.id, 'clicked_at')).toBe(4);
    expect(Number(sql(`SELECT count(*) FROM publication_banner_seen WHERE user_id='${p.id}' AND clicked_at IS NOT NULL`))).toBe(4);
  });

  test('VIP banner only for tagged players, the server refuses counting a banner the player may not see, a console change shows up', async ({ page }) => {
    const p = await start(page, 'vip', { lang: 'fa' });
    expect(await slideKeys(page)).not.toContain('star_bn_vip');
    const vipId = sql(`SELECT id FROM publication_banner WHERE key='star_bn_vip'`);
    const view = await gameRpc(page, 'publication_banner_view', { bannerIds: [vipId], platform: 'web' });
    expect(view.ok && view.data.counted).toEqual([]);
    const click = await gameRpc(page, 'publication_banner_click', { bannerId: vipId, platform: 'web' });
    expect(click).toMatchObject({ ok: false, code: 'BANNER_NOT_AVAILABLE' });
    expect(Number(sql(`SELECT count(*) FROM publication_banner_seen WHERE user_id='${p.id}' AND banner_id='${vipId}'`))).toBe(0);

    const tag = sql(`SELECT id FROM base_tag WHERE key_type='user' AND key='star-vip'`);
    await base.put('/players/' + p.id + '/tags', { tagIds: [tag] });
    const list = await pub.get('/banner-lists?key=menu_carousel');
    const urlBanner = await pub.post('/banners', { key: 'e2e_' + k + '_url', listId: list.items[0].id, priority: 500, title: { fa: 'خبر داغ ' + k, en: 'Hot news ' + k },
      action: { type: 'url', url: { fa: 'https://example.com/fa', en: 'https://example.com/en' } } });
    extra.push(urlBanner.id);
    await page.reload();
    await waitHome(page);
    const keys = await slideKeys(page);
    expect(keys[0]).toBe('e2e_' + k + '_url');
    expect(keys).toContain('star_bn_vip');
    const hot = await showSlide(page, 'e2e_' + k + '_url');
    const popup = page.waitForEvent('popup');
    await hot.click();
    expect((await popup).url()).toContain('example.com/fa');

    await (await showSlide(page, 'star_bn_vip')).click();
    await screen(page, 'shop');
    await expect(page.locator('#products [data-key="star_pack_vip"]')).toHaveClass(/bn-flash/);

    await pub.put('/banners/' + urlBanner.id, { key: 'e2e_' + k + '_url', listId: list.items[0].id, priority: 500, isActive: false, title: { fa: 'خبر داغ ' + k }, action: { type: 'none' } });
    await page.reload();
    await waitHome(page);
    await expect(carousel(page).locator('.bn-slide').first()).toBeVisible();
    expect(await slideKeys(page)).not.toContain('e2e_' + k + '_url');
    const stale = await gameRpc(page, 'publication_banner_click', { bannerId: urlBanner.id });
    expect(stale).toMatchObject({ ok: false, code: 'BANNER_NOT_AVAILABLE' });
  });

  test('English at 390px: English titles and image, Latin digits, no sideways scroll', async ({ spawn }) => {
    const m = await spawn({ mobile: true, signedIn: false });
    await start(m, 'en', { lang: 'en' });
    const first = carousel(m).locator('.bn-slide').first();
    await expect(first).toBeVisible();
    await expect(first).toContainText('Welcome to Star Catcher');
    await expect(first.locator('img')).toHaveAttribute('src', /welcome-en\.svg$/);
    await expect(carousel(m).locator('.bn-dot').nth(1)).toHaveAttribute('aria-label', 'Banner 2 of 4');
    const box = await carousel(m).boundingBox();
    expect(box!.width).toBeLessThanOrEqual(390);
    await noHorizontalScroll(m);
  });
});
