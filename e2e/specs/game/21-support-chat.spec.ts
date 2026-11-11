import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { base, deleteUser, uid } from '../../lib/api';
import { root, settle, toasts } from '../../lib/console';
import { CONSOLE_URL, FIXTURES } from '../../lib/env';
import { openGame, screen, toast } from '../../lib/game';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
});

async function start(page: Page, tag: string) {
  const p = await openGame(page, `e2e-sc-${tag}-${k}`);
  players.push(p.id);
  return p;
}

async function agent(spawn: (o?: any) => Promise<Page>, hash: string) {
  const c = await spawn({ signedIn: true });
  await c.goto(CONSOLE_URL + '/' + hash);
  await c.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await root(c).waitFor();
  await settle(c);
  return c;
}

test.describe('game: support', () => {
  test('a player sends a ticket with a picture, gets the answer live, solves it and rates it', async ({ page, spawn }) => {
    const me = await start(page, 't');
    await page.locator('#menu-support').click();
    await screen(page, 'support');
    await expect(page.locator('#ticket-dept option')).toHaveCount(2);
    await page.locator('#ticket-dept').selectOption('payments');
    await expect(page.locator('#ticket-dept-hint')).toContainText('پرداخت ناموفق');
    await page.locator('#ticket-send').click();
    await expect(page.locator('#ticket-msg')).toHaveText('موضوع را بنویس (حداکثر ۱۵۰ حرف).');
    await page.locator('#ticket-subject').fill('سکه‌ها نرسید ' + k);
    await page.locator('#ticket-body').fill('پرداخت کردم ولی سکه نیامد.');
    await page.locator('#ticket-file').setInputFiles(path.join(FIXTURES, 'e2e-shot.png'));
    await expect(page.locator('#ticket-files .file-chip')).toContainText('e2e-shot.png');
    await page.locator('#ticket-send').click();
    await screen(page, 'ticket');
    await expect(page.locator('#ticket-title')).toHaveText('سکه‌ها نرسید ' + k);
    await expect(page.locator('#ticket-meta [data-status="open"]')).toHaveText('ثبت شد');
    await expect(page.locator('#ticket-thread [data-attachment] img')).toBeVisible();
    const number = await page.locator('#ticket-number').innerText();

    const c = await agent(spawn, '#/support/tickets?q=' + encodeURIComponent(number));
    await root(c).locator('tbody tr.x-table-row', { hasText: number }).click();
    const tk = root(c).locator('[data-ticket]');
    await expect(tk.locator('[data-attachment] img')).toBeVisible();
    await tk.locator('[data-composer]').fill('سلام، سکه‌هایت را اضافه کردیم.');
    await tk.locator('[data-send]').click();
    await expect(toasts(c).last()).toContainText('فرستاده شد');

    await expect(page.locator('#ticket-thread .chat-msg.staff')).toContainText('سکه‌هایت را اضافه کردیم', { timeout: 20_000 });
    await expect(page.locator('#ticket-meta [data-status]')).toHaveText('منتظر جواب تو');
    await page.locator('#ticket-reply-body').fill('رسید، ممنون');
    await page.locator('#ticket-reply-send').click();
    await expect(page.locator('#ticket-thread .chat-msg.mine').last()).toContainText('رسید، ممنون');
    await page.locator('#ticket-resolve').click();
    await expect(toast(page)).toHaveText('خوشحالیم که مشکلت حل شد.');
    await expect(page.locator('#ticket-meta [data-status]')).toHaveText('حل شد');
    await page.locator('#ticket-rate [data-star="5"]').click();
    await page.locator('#ticket-rate [data-rate-send]').click();
    await expect(page.locator('#ticket-rate [data-rated="5"]')).toHaveText('امتیاز تو: ۵ از ۵');
    const t2 = (await base.get('/support/tickets?q=' + number)).items[0];
    expect(t2).toMatchObject({ status: 'resolved', rating: 5, userId: me.id });

    await page.locator('#screen-ticket [data-go="support"]').click();
    await expect(page.locator(`#ticket-list [data-ticket]`).first()).toContainText('حل شد');
  });

  test('a player and a console agent chat live: queue, take, typing on both sides, messages, then the agent closes it into a ticket', async ({ page, spawn }) => {
    const me = await start(page, 'lc');
    const c = await agent(spawn, '#/support/live-chat');
    await c.locator('[data-lc-tabs] [data-value="queue"]').click();
    await page.locator('#menu-support').click();
    await screen(page, 'support');
    await expect(page.locator('#live-state')).toContainText('آنلاین است');
    await page.locator('#btn-livechat').click();
    await screen(page, 'livechat');
    await page.locator('#lc-input').fill('سلام، کسی هست؟ ' + k);
    await page.locator('#lc-send').click();
    await expect(page.locator('#lc-thread .chat-msg.mine')).toContainText('کسی هست');
    await expect(page.locator('#lc-banner')).toContainText('در صف هستی');

    const item = c.locator(`[data-chat-status="waiting"]`, { hasText: me.username });
    await expect(item).toBeVisible({ timeout: 15_000 });
    await item.locator('[data-take]').click();
    await expect(toasts(c).last()).toContainText('گفتگو مال شما شد');
    const pane = c.locator('[data-lc-pane]');
    await expect(pane.locator('[data-thread]')).toContainText('کسی هست؟ ' + k);
    await expect(page.locator('#lc-state')).toContainText('در حال گفتگو با', { timeout: 15_000 });

    await pane.locator('[data-composer]').pressSequentially('در خدمت');
    await expect(page.locator('#lc-typing')).toBeVisible({ timeout: 10_000 });
    await pane.locator('[data-composer]').fill('در خدمتم، بفرمایید.');
    await pane.locator('[data-send]').click();
    await expect(page.locator('#lc-thread .chat-msg.staff')).toContainText('در خدمتم، بفرمایید.', { timeout: 10_000 });

    await page.locator('#lc-input').pressSequentially('یک سؤال');
    await expect(pane.locator('[data-typing]')).toBeVisible({ timeout: 20_000 });
    await page.locator('#lc-send').click();
    await expect(pane.locator('[data-thread]')).toContainText('یک سؤال', { timeout: 20_000 });
    await expect(pane.locator('[data-player-online="true"]')).toBeVisible();

    await pane.locator('[data-close-chat]').click();
    await c.locator('[data-form-dialog="chat-close"] [data-to-ticket]').click();
    await c.locator('[data-form-dialog="chat-close"] [data-dialog-ok]').click();
    await expect(toasts(c).last()).toContainText('تیکت');
    await expect(pane.locator('[data-chat-ticket]')).toBeVisible();

    await expect(page.locator('#lc-state')).toHaveText('گفتگو تمام شد.', { timeout: 15_000 });
    await expect(page.locator('#lc-form')).toBeHidden();
    await page.locator('#lc-actions [data-star="4"]').click();
    await page.locator('#lc-actions [data-rate-send]').click();
    await expect(page.locator('#lc-actions [data-rated="4"]')).toBeVisible();
    await page.locator('#lc-actions button', { hasText: 'گفتگو به تیکت' }).click();
    await screen(page, 'ticket');
    await expect(page.locator('#ticket-thread')).toContainText('یک سؤال');
    const chats = await base.get('/support/chats?view=closed&userId=' + me.id);
    expect(chats.items[0]).toMatchObject({ status: 'closed', rating: 4 });
  });
});
