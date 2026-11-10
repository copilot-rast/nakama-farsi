import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { base, deleteUser, uid } from '../../lib/api';
import { sql } from '../../lib/db';
import { openGame, screen, toast } from '../../lib/game';

const k = uid();
const room = 'e2e-' + k.replace(/_/g, '-').toLowerCase();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  sql(`DELETE FROM base_chat_report WHERE channel_label LIKE 'e2e%'; DELETE FROM message WHERE stream_label LIKE 'e2e%';`);
});

async function start(page: Page, tag: string) {
  const p = await openGame(page, `e2e-chat-${tag}-${k}`, { query: '?room=' + room });
  await expect(page.locator('#screen-home')).toBeVisible({ timeout: 20_000 });
  players.push(p.id);
  return p;
}

async function toChat(page: Page) {
  await page.locator('#menu-chat').click();
  await screen(page, 'chat');
  await expect(page.locator('#chat-list .empty, #chat-list .chat-msg').first()).toBeVisible();
  await expect(page.locator('#chat-list')).not.toContainText('در حال ورود');
}

async function say(page: Page, text: string) {
  await page.locator('#chat-input').fill(text);
  await page.locator('#chat-send').click();
}

test.describe('game: chat', () => {
  test('two players chat live, bad words are masked, a banned word is refused, a message is reported with a reason', async ({ page, spawn }) => {
    const a = await start(page, 'a');
    const other = await spawn({ signedIn: false });
    const b = await start(other, 'b');
    await toChat(page);
    await toChat(other);
    await say(page, 'سلام ' + k + '، تو اح‌ـــمق هستی؟');
    const mine = page.locator('#chat-list .chat-msg.mine').last();
    await expect(mine).toContainText('تو **** هستی؟');
    await expect(mine.locator('.chat-report')).toHaveCount(0);
    const seen = other.locator('#chat-list .chat-msg:not(.mine)').last();
    await expect(seen).toContainText('تو **** هستی؟');
    await expect(seen).toContainText(a.username);
    await expect(page.locator('#chat-input')).toHaveValue('');

    await say(page, 'تو حرام‌زاده‌ای');
    await expect(toast(page)).toHaveText('این پیام کلمهٔ ممنوع داشت و فرستاده نشد.');
    await expect(page.locator('#chat-input')).toHaveValue('تو حرام‌زاده‌ای');
    await expect(other.locator('#chat-list')).not.toContainText('حرام');

    await seen.locator('.chat-report').click();
    const dlg = other.locator('#dialog');
    await expect(dlg).toContainText('چرا این پیام را گزارش می‌کنی؟');
    await expect(dlg.locator('.report-quote')).toContainText('تو **** هستی؟');
    await dlg.locator('#report-send').click();
    await expect(dlg.locator('.form-msg')).toHaveText('یک دلیل انتخاب کن.');
    await dlg.locator('.report-reason', { hasText: 'توهین یا فحاشی' }).click();
    await dlg.locator('#report-note').fill('e2e فحش داد');
    await dlg.locator('#report-send').click();
    await expect(toast(other)).toHaveText('گزارشت ثبت شد. ممنونیم!');
    const reports = await base.get('/reports?source=player&senderId=' + a.id);
    expect(reports.total).toBe(1);
    expect(reports.items[0]).toMatchObject({ reporterId: b.id, titleKey: 'abuse', note: 'e2e فحش داد', channelLabel: room });

    await other.locator('#chat-list .chat-msg:not(.mine)').last().locator('.chat-report').click();
    await other.locator('#dialog .report-reason', { hasText: 'هرزنامه یا تبلیغ' }).click();
    await other.locator('#dialog #report-send').click();
    await expect(toast(other)).toHaveText('این پیام را قبلاً گزارش کرده‌ای.');
  });

  test('a muted player sees why and until when at once, cannot send, and can chat again when the mute is lifted', async ({ page }) => {
    const a = await start(page, 'm');
    await toChat(page);
    await say(page, 'سلام ' + k);
    await expect(page.locator('#chat-list .chat-msg.mine').last()).toContainText('سلام');
    const r = await base.post('/players/' + a.id + '/restrictions', { kind: 'mute', minutes: 90, reason: { fa: 'e2e کمی آرام‌تر', en: 'e2e calm down' } });
    const banner = page.locator('#chat-banner');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText('فعلاً بی‌صدا هستی');
    await expect(banner).toContainText('e2e کمی آرام‌تر');
    await expect(banner).toContainText('نمی‌توانی در گفتگو پیام بفرستی');
    await expect(page.locator('#chat-input')).toBeDisabled();
    await page.reload();
    await expect(page.locator('#screen-home')).toBeVisible({ timeout: 20_000 });
    await toChat(page);
    await expect(page.locator('#chat-banner')).toContainText('e2e کمی آرام‌تر');
    await expect(page.locator('#chat-send')).toBeDisabled();
    await base.post('/restrictions/' + r.id + '/lift', { note: 'e2e' });
    await expect(toast(page)).toHaveText('حالا دوباره می‌توانی پیام بفرستی.');
    await expect(page.locator('#chat-banner')).toBeHidden();
    await say(page, 'برگشتم ' + k);
    await expect(page.locator('#chat-list .chat-msg.mine').last()).toContainText('برگشتم');
  });

  test('a warning pops up; a suspended player is signed out with the reason and end time, stays out after a reload, and is back after the lift', async ({ page }) => {
    const a = await start(page, 's');
    await base.post('/players/' + a.id + '/restrictions', { kind: 'warn', reason: { fa: 'e2e لطفاً مؤدب باش', en: 'e2e be nice' } });
    await expect(page.locator('#dialog')).toContainText('اخطار');
    await expect(page.locator('#dialog')).toContainText('e2e لطفاً مؤدب باش');
    await page.locator('#dialog .btn-primary').click();
    const r = await base.post('/players/' + a.id + '/restrictions', { kind: 'suspend', minutes: 60 * 24, reason: { fa: 'e2e یک روز استراحت', en: 'e2e one day off' } });
    await screen(page, 'suspended');
    await expect(page.locator('#screen-suspended h1')).toHaveText('حسابت موقتاً بسته است');
    await expect(page.locator('#suspended-reason')).toHaveText('e2e یک روز استراحت');
    await expect(page.locator('#suspended-until')).toContainText('نمی‌توانی وارد بازی شوی');
    await page.reload();
    await screen(page, 'suspended');
    await expect(page.locator('#suspended-reason')).toHaveText('e2e یک روز استراحت');
    await expect(page.locator('#screen-home')).toBeHidden();
    await base.post('/restrictions/' + r.id + '/lift', { note: 'e2e' });
    await page.locator('#btn-suspended-retry').click();
    await expect(page.locator('#screen-home')).toBeVisible({ timeout: 20_000 });
  });

  test('a public player tag shows on home and in the account, and unlocks the VIP pack in the shop', async ({ page, spawn }) => {
    const a = await start(page, 't');
    const vip = (await base.get('/tags?keyType=user&q=star-vip')).items.find((t: any) => t.key === 'star-vip');
    expect(vip.isPublic).toBe(true);
    await base.put('/players/' + a.id + '/tags', { tagIds: [vip.id] });
    await page.reload();
    await expect(page.locator('#screen-home')).toBeVisible({ timeout: 20_000 });
    await expect(page.locator('#my-tags [data-tag="star-vip"]')).toHaveText('عضو ویژه');
    await page.locator('#menu-account').click();
    await expect(page.locator('#account-tags [data-tag="star-vip"]')).toHaveText('عضو ویژه');
    await page.locator('#screen-account [data-go="home"]').click();
    await page.locator('[data-go="shop"]').first().click();
    const pack = page.locator('.product[data-key="star_pack_vip"]');
    await expect(pack).toBeVisible();
    await expect(pack.locator('.badge-exclusive')).toHaveText('ویژهٔ تو');
    const other = await spawn({ signedIn: false });
    await start(other, 'u');
    await expect(other.locator('#my-tags')).toBeHidden();
    await other.locator('[data-go="shop"]').first().click();
    await expect(other.locator('.product').first()).toBeVisible();
    await expect(other.locator('.product[data-key="star_pack_vip"]')).toHaveCount(0);
  });
});
