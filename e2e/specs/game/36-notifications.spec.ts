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
  const p = await openGame(page, `e2e-nt-${tag}-${k}`, opts);
  players.push(p.id);
  return p;
}

async function categoryId(key: string) {
  return (await pub.get('/notify-categories?key=' + key)).items[0].id;
}

test.describe('notifications: Star Catcher', () => {
  test('settings screen: this device is registered, categories switch off and back on, the device can be turned off', async ({ page }) => {
    const p = await start(page, 'set', { lang: 'fa' });
    await expect.poll(() => Number(sql(`SELECT count(*) FROM publication_push_device WHERE user_id='${p.id}' AND is_active AND platform='web' AND token LIKE 'sandbox-web-%'`))).toBe(1);
    await page.locator('#screen-home [data-go="notify"]').click();
    await screen(page, 'notify');
    await expect(page.locator('.nt-device-state')).toHaveAttribute('data-push', 'on');
    await expect(page.locator('[data-push-token]')).toHaveText(/^sandbo…[0-9a-f]{4}$/);
    const system = page.locator('.nt-pref', { hasText: 'پیام‌های مهم' });
    await expect(system).toContainText('همیشه روشن');
    await expect(system.locator('input')).toHaveCount(0);
    const offers = page.locator('[data-pref="offers"]');
    await expect(offers).toBeChecked();
    await page.locator('label[for="pref-offers"]').click();
    await expect(page.locator('#toast')).toContainText('پیشنهادها و تخفیف‌ها: خاموش');
    await expect(offers).not.toBeChecked();
    expect(sql(`SELECT p.enabled FROM publication_notify_pref p JOIN publication_notify_category c ON c.id=p.category_id WHERE p.user_id='${p.id}' AND c.key='offers'`)).toBe('f');
    await page.reload();
    await page.locator('#screen-home [data-go="notify"]').click();
    await expect(page.locator('[data-pref="offers"]')).not.toBeChecked();
    await page.locator('#btn-push-toggle').click();
    await expect(page.locator('.nt-device-state')).toHaveAttribute('data-push', 'off');
    expect(sql(`SELECT count(*) FROM publication_push_device WHERE user_id='${p.id}'`)).toBe('0');
    await page.locator('#btn-push-toggle').click();
    await expect(page.locator('.nt-device-state')).toHaveAttribute('data-push', 'on');
    await page.locator('label[for="pref-offers"]').click();
    await expect(page.locator('[data-pref="offers"]')).toBeChecked();
  });

  test('a console campaign reaches the inbox with its picture and category, a tap opens the shop and counts as opened; an opted-out category is skipped', async ({ page }) => {
    const p = await start(page, 'inbox', { lang: 'fa' });
    await expect.poll(() => Number(sql(`SELECT count(*) FROM publication_push_device WHERE user_id='${p.id}' AND is_active`))).toBe(1);
    const off = await gameRpc(page, 'publication_notify_prefs_set', { prefs: { offers: false } });
    expect(off.ok).toBe(true);
    const skipped = await pub.post('/notifications', { channel: 'push_in_app', title: { fa: 'تخفیف ' + k }, body: { fa: 'x' }, categoryId: await categoryId('offers'), target: { mode: 'users', userIds: [p.id] } });
    expect(skipped).toMatchObject({ recipientCount: 1, skippedCount: 1 });
    const n = await pub.post('/notifications', { channel: 'push_in_app', title: { fa: 'جایزهٔ ویژه ' + k, en: 'Special prize ' + k }, body: { fa: '**صد سکه** منتظر توست' }, imageUrl: '/stargame/banners/nova.svg',
      categoryId: await categoryId('rewards'), action: { type: 'goto', value: 'shop', label: { fa: 'برو به فروشگاه' } }, target: { mode: 'users', userIds: [p.id] } });
    await expect(page.locator('#note-toast')).toContainText('جایزهٔ ویژه ' + k, { timeout: 25_000 });
    await expect(page.locator('#inbox-badge')).toHaveText('۱');
    await page.locator('#btn-inbox').click();
    await screen(page, 'inbox');
    const item = page.locator(`.inbox-item[data-inbox="${n.id}"]`);
    await expect(item).toHaveClass(/unread/);
    await expect(item.locator('img.inbox-img')).toHaveAttribute('src', /nova\.svg$/);
    await expect(item.locator('[data-inbox-category="rewards"]')).toHaveText('جایزه‌ها و دستاوردها');
    await expect(page.locator('.inbox-item', { hasText: 'تخفیف ' + k })).toHaveCount(0);
    await item.locator('button[data-action="goto"]').click();
    await screen(page, 'shop');
    await expect.poll(async () => (await pub.get('/notifications/' + n.id)).openedCount).toBe(1);
    expect((await pub.get('/notifications/' + n.id)).pushSent).toBe(1);
  });

  test('English settings and inbox at 390px', async ({ spawn }) => {
    const m = await spawn({ mobile: true });
    await start(m, 'en', { lang: 'en' });
    await m.locator('#screen-home [data-go="notify"]').click();
    await screen(m, 'notify');
    await expect(m.locator('#screen-notify h1')).toHaveText('Notifications');
    await expect(m.locator('.nt-pref', { hasText: 'Important messages' })).toContainText('Always on');
    await expect(m.locator('.nt-device-state')).toHaveText('Notifications are on for this device.');
    await noHorizontalScroll(m);
  });
});
