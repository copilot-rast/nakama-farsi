import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { base, deleteUser, eco, uid } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { balance, gameRpc, openGame, persian, screen, toast, waitHome } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  restoreConfig();
});

async function start(page: Page, tag: string, lang?: string) {
  const p = await openGame(page, `e2e-gg-${tag}-${k}`, lang ? { lang } : {});
  players.push(p.id);
  return p;
}

test.describe('game: remote config', () => {
  test('the home banner and the golden rush switch follow the console config, per player tag', async ({ page }) => {
    const me = await start(page, 'cfg');
    const banner = page.locator('#config-banner');
    await expect(banner).toHaveText('به ستاره‌گیر خوش آمدی! هر روز سر بزن و جایزه بگیر.');
    await expect(banner).toHaveAttribute('data-variant', 'default');
    await expect(page.locator('#rush-badge')).toBeVisible();
    expect(await page.evaluate(() => (window as any).eval('goldenRush()'))).toBe(true);

    const vip = (await base.get('/tags?keyType=user&q=star-vip')).items.find((t: any) => t.key === 'star-vip');
    await base.put('/players/' + me.id + '/tags', { tagIds: [vip.id] });
    const rush = (await base.get('/remote-config?q=stargame.golden_rush')).items[0];
    await base.put('/remote-config/' + rush.id, { key: rush.key, type: 'bool', description: rush.description, value: false, rules: [], isActive: true });
    await page.reload();
    await waitHome(page);
    await expect(banner).toHaveText('عضو ویژه، امروز سکه‌های بیشتری منتظرت است!');
    await expect(banner).toHaveAttribute('data-variant', 'vip');
    await expect(page.locator('#rush-badge')).toBeHidden();
    expect(await page.evaluate(() => (window as any).eval('goldenRush()'))).toBe(false);

    const acts = await gameRpc(page, 'base_my_activity', { kind: 'signin' });
    expect(acts.ok).toBe(true);
    expect(acts.data.items.length).toBeGreaterThan(0);
    restoreConfig();
  });
});

test.describe('game: referral code and my activity', () => {
  test('a friend enters my code, both get coins after the first round, and my activity shows it', async ({ page, spawn }) => {
    const rules = await eco.get('/referral-settings');
    await eco.put('/referral-settings', { ...rules, holdFlagged: false });
    await start(page, 'ra');
    await page.locator('#menu-referral').click();
    await screen(page, 'referral');
    await expect(page.locator('#ref-code')).not.toHaveText('…');
    const code = (await page.locator('#ref-code').innerText()).trim();
    expect(code).toMatch(/^[A-Z2-9]{8}$/);
    await expect(page.locator('#ref-rewards')).toHaveText('تو ۱۰۰ سکه می‌گیری و دوستت ۵۰ سکه.');
    await expect(page.locator('#ref-goal')).toContainText('پایان یک دور');
    await page.locator('#ref-input').fill(code);
    await page.locator('#ref-send').click();
    await expect(page.locator('#ref-msg')).toHaveText('نمی‌توانی کد خودت را وارد کنی.');

    const friendPage = await spawn({ signedIn: false });
    const friend = await start(friendPage, 'rb');
    await friendPage.locator('#menu-referral').click();
    await screen(friendPage, 'referral');
    await friendPage.locator('#ref-input').fill('NOPE2345');
    await friendPage.locator('#ref-send').click();
    await expect(friendPage.locator('#ref-msg')).toHaveText('این کد پیدا نشد. حروف را دوباره نگاه کن.');
    await friendPage.locator('#ref-input').fill(persian(code.toLowerCase()));
    await friendPage.locator('#ref-send').click();
    await expect(friendPage.locator('#ref-msg')).toHaveText('کد ثبت شد. جایزه‌ها بعد از کامل کردن شرط می‌رسد.');
    await expect(friendPage.locator('#ref-form')).toBeHidden();
    await expect(friendPage.locator('#ref-by [data-referred-by]')).toContainText('تو با کد');

    const before = await balance(page);
    const fBefore = await balance(friendPage);
    await eco.post('/users/' + friend.id + '/events', { key: 'round_finished', value: 5 });
    await page.reload();
    await waitHome(page);
    await expect.poll(() => balance(page)).toBe(before + 100);
    await friendPage.reload();
    await waitHome(friendPage);
    await expect.poll(() => balance(friendPage)).toBe(fBefore + 50);

    await page.locator('#menu-referral').click();
    await expect(page.locator('#ref-counts')).toHaveText('۱ دوست با کد تو آمده‌اند؛ ۱ نفر شرط را کامل کرده‌اند.');
    await expect(page.locator('#ref-recent li')).toContainText(friend.username);
    await page.locator('#screen-referral [data-go="home"]').click();

    await page.locator('#menu-activity').click();
    await screen(page, 'activity');
    await expect(page.locator('#act-list [data-act="referral:joined"]')).toContainText(friend.username + ' با کد تو آمد');
    await expect(page.locator('#act-list [data-act="prize:granted"]').first()).toBeVisible();
    await expect(page.locator('#act-list [data-act="signin:device"]').first()).toHaveText(/وارد بازی شدی/);
    await page.locator('#act-kinds [data-kind="referral"]').click();
    await expect(page.locator('#act-list [data-act^="signin"]')).toHaveCount(0);
    await expect(page.locator('#act-list [data-act^="referral"]').first()).toBeVisible();
    await expect(toast(page)).toBeHidden();
  });
});
