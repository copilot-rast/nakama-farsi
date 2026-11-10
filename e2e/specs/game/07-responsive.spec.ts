import { expect, test } from '../../lib/fixtures';
import { deleteUser, uid } from '../../lib/api';
import { noHorizontalScroll } from '../../lib/console';
import { openGame, screen } from '../../lib/game';

test.use({ pageOpts: { signedIn: false, mobile: true } } as any);

const players: string[] = [];
test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

test('game screens fit a 390px phone', async ({ page }) => {
  const me = await openGame(page, 'e2e-g-mobile-' + uid());
  players.push(me.id);
  await noHorizontalScroll(page);
  const menu = await page.locator('.menu-grid').boundingBox();
  expect(menu!.width).toBeLessThanOrEqual(390);
  for (const name of ['shop', 'gift', 'payments', 'news', 'faq']) {
    await page.locator(`#screen-home [data-go="${name}"]`).click();
    await screen(page, name);
    await expect(page.locator(`#screen-${name} .spinner, #screen-${name} .skel`)).toHaveCount(0);
    await noHorizontalScroll(page);
    await page.locator(`#screen-${name} [data-go="home"]`).first().click();
    await screen(page, 'home');
  }
  await page.locator('#btn-inbox').click();
  await screen(page, 'inbox');
  await noHorizontalScroll(page);
  await page.locator('#screen-inbox [data-go="home"]').click();
  await page.locator('#btn-play').click();
  await screen(page, 'play');
  await page.waitForFunction(() => (window as any).__stargame.phase === 'countdown' || (window as any).__stargame.phase === 'playing');
  const canvas = await page.locator('#canvas').boundingBox();
  expect(canvas!.width).toBeLessThanOrEqual(390);
  expect(canvas!.width).toBeGreaterThan(300);
  await noHorizontalScroll(page);
  const tap = await page.evaluate(() => (window as any).__stargame.width);
  expect(tap).toBeGreaterThan(300);
  await page.locator('#btn-quit').click();
  await screen(page, 'home');
});
