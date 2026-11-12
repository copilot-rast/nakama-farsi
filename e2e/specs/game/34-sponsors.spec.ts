import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, pub, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll } from '../../lib/console';
import { gameRpc, openGame, screen, type GamePlayer } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string, opts: { lang?: string } = {}): Promise<GamePlayer> {
  const p = await openGame(page, `e2e-sp-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

function seen(userId: string, col: 'viewed_at' | 'clicked_at') {
  return Number(sql(`SELECT count(*) FROM publication_sponsor_seen s JOIN publication_sponsor p ON p.id=s.sponsor_id WHERE s.user_id='${userId}' AND s.${col} IS NOT NULL AND p.key='star_sp_nova'`));
}

async function openLuck(page: Page) {
  await page.locator('#screen-home [data-go="chance"]').click();
  await screen(page, 'chance');
  const box = page.locator('#sponsor-box');
  await expect(box).toBeVisible();
  return box;
}

test.describe('sponsors: Star Catcher', () => {
  test('luck screen: the sponsored wheel banner with the sponsor credit, one view a day, the card and the links open the sponsor page', async ({ page }) => {
    const p = await start(page, 'fa', { lang: 'fa' });
    const box = await openLuck(page);
    const card = box.locator('.sp-card[data-banner="star_bn_nova"]');
    await expect(card).toContainText('چرخ روزانه با حمایت آبمیوهٔ نوا');
    await expect(card.locator('[data-sponsor-credit="star_sp_nova"]')).toHaveText('با حمایت آبمیوهٔ نوا');
    await expect(card.locator('.sp-logo')).toHaveAttribute('src', /nova-logo\.svg$/);
    await expect(card.locator('.bn-cta')).toHaveText('دربارهٔ نوا');
    await expect(box.locator('[data-sponsor-link]')).toHaveCount(2);
    await expect(box.locator('[data-sponsor-link="telegram"]')).toHaveText('کانال تلگرام نوا');
    await expect.poll(() => seen(p.id, 'viewed_at')).toBe(1);

    await page.locator('#screen-chance [data-go="home"]').first().click();
    await openLuck(page);
    await page.reload();
    await openLuck(page);
    await page.waitForTimeout(800);
    expect(seen(p.id, 'viewed_at')).toBe(1);
    expect(Number(sql(`SELECT count(*) FROM publication_banner_seen s JOIN publication_banner b ON b.id=s.banner_id WHERE s.user_id='${p.id}' AND b.key='star_bn_nova' AND s.viewed_at IS NOT NULL`))).toBe(1);

    const popup = page.waitForEvent('popup');
    await page.locator('#sponsor-box .sp-card').click();
    const site = await popup;
    expect(site.url()).toMatch(/\/stargame\/sponsor\.html$/);
    await expect(site.locator('h1')).toHaveText('آبمیوهٔ نوا');
    await expect.poll(() => seen(p.id, 'clicked_at')).toBe(1);

    const tg = page.waitForEvent('popup');
    await page.locator('#sponsor-box [data-sponsor-link="telegram"]').click();
    expect((await tg).url()).toContain('sponsor.html?via=telegram');
    expect(seen(p.id, 'clicked_at')).toBe(1);

    const list = await gameRpc(page, 'publication_sponsors', { lang: 'fa' });
    const nova = list.data.items.find((x: any) => x.key === 'star_sp_nova');
    expect(nova).toMatchObject({ name: 'آبمیوهٔ نوا', color: '#0f766e' });
    expect(nova.contactEmail).toBeUndefined();
    expect(nova.notes).toBeUndefined();
  });

  test('a sponsor switched off in the console hides the card; English with the credit at 390px', async ({ page, spawn }) => {
    const sp = (await pub.get('/sponsors?key=star_sp_nova')).items[0];
    const off = await pub.put('/sponsors/' + sp.id, { ...sp, isActive: false });
    expect(off.isActive).toBe(false);
    try {
      await start(page, 'off', { lang: 'fa' });
      await page.locator('#screen-home [data-go="chance"]').click();
      await screen(page, 'chance');
      await page.waitForTimeout(800);
      await expect(page.locator('#sponsor-box')).toBeHidden();
      const res = await gameRpc(page, 'publication_banners', { placement: 'luck_sponsor', platform: 'web' });
      expect(res.data.items).toEqual([]);
    } finally {
      await pub.put('/sponsors/' + sp.id, { ...sp, isActive: true });
    }

    const m = await spawn({ mobile: true });
    await start(m, 'en', { lang: 'en' });
    const box = await openLuck(m);
    await expect(box.locator('[data-sponsor-credit="star_sp_nova"]')).toHaveText('Sponsored by Nova Juice');
    await expect(box.locator('.sp-card')).toContainText('The daily wheel, sponsored by Nova Juice');
    await expect(box.locator('[data-sponsor-link="instagram"]')).toHaveText('Nova on Instagram');
    await noHorizontalScroll(m);
  });
});
