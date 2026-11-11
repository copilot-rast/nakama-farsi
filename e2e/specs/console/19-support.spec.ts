import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, deleteUser, devicePlayer, rpc, uid, type Player } from '../../lib/api';
import { field, fillTranslated, open, openSettings, panel, psWindow, root, rows, savePanel, settle, settlePanel, toasts } from '../../lib/console';
import { restoreConfig } from '../../lib/db';
import { CONSOLE_USER } from '../../lib/env';
import { PNG, bearerOf, dropConsoleUser, makeConsoleUser, signIn, uploadAs } from '../../lib/support';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  restoreConfig();
});

async function newPlayer(tag: string): Promise<Player> {
  const p = await devicePlayer(`e2e-sup-${tag}-${k}`);
  players.push(p.id);
  return p;
}

test.describe('support: settings sections', () => {
  test('departments and canned replies are edited in project settings and show up for players and agents', async ({ page }) => {
    await openSettings(page, 'support-departments');
    await expect(rows(page).filter({ hasText: 'پشتیبانی عمومی' })).toHaveCount(1);
    await expect(rows(page).filter({ hasText: 'خرید و پرداخت' })).toContainText('۱۲ ساعت');
    await root(page).getByRole('button', { name: 'بخش تازه' }).first().click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k);
    await fillTranslated(p, 'name', { fa: 'فنی ' + k, en: 'Tech ' + k });
    await field(p, 'مهلت اولین پاسخ').locator('input').fill('۶');
    await field(p, 'پشتیبان‌های این بخش').getByText(CONSOLE_USER, { exact: true }).click();
    await savePanel(page);
    await expect(rows(page).filter({ hasText: 'فنی ' + k })).toContainText('۶ ساعت');
    const pl = await newPlayer('dept');
    const info = await rpc(pl, 'base_support_info', { lang: 'en' });
    expect(info.data.departments.map((d: any) => d.name)).toContain('Tech ' + k);

    await openSettings(page, 'support-canned');
    await expect(rows(page).filter({ hasText: 'سلام و خوش‌آمد' })).toHaveCount(1);
    await root(page).getByRole('button', { name: 'پاسخ آمادهٔ تازه' }).first().click();
    await settlePanel(page);
    await field(panel(page), 'کلید').locator('input').fill('e2e_' + k);
    await fillTranslated(panel(page), 'title', { fa: 'بازپرداخت ' + k });
    await expect(panel(page).locator('.eco-invalid')).toHaveCount(0);
    await panel(page).locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(panel(page).locator('.eco-invalid').first()).toBeVisible();
    await fillTranslated(panel(page), 'body', { fa: 'متن اول ' + k });
    await savePanel(page);
    await root(page).getByRole('button', { name: 'پاسخ آمادهٔ تازه' }).first().click();
    await settlePanel(page);
    await field(panel(page), 'کلید').locator('input').fill('e2e_' + k + 'b');
    await fillTranslated(panel(page), 'title', { fa: 'پیگیری ' + k });
    await fillTranslated(panel(page), 'body', { fa: 'پرداختت را پیگیری کردیم ' + k, en: 'We checked your payment ' + k });
    await savePanel(page);
    await expect(rows(page).filter({ hasText: 'پیگیری ' + k })).toContainText('همهٔ بخش‌ها');

    await openSettings(page, 'support-settings');
    const form = psWindow(page).locator('[data-settings-form="support-settings"]');
    await field(form, 'حداکثر پیوست هر پیام').locator('input').fill('۳');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    expect((await base.get('/support/settings')).maxAttachments).toBe(3);
    restoreConfig();
  });
});

test.describe('support: tickets', () => {
  test('a player ticket with a picture is answered with a canned reply, a hidden note, assignment and rating', async ({ page }) => {
    await base.post('/support/canned', { key: 'e2e_' + k + 'c', title: { fa: 'پیگیری ' + k }, body: { fa: 'پرداختت را پیگیری کردیم ' + k, en: 'We checked your payment ' + k }, isActive: true, sort: 1 });
    const pl = await newPlayer('t');
    const shot = await uploadAs(pl, PNG, 'screenshot.png');
    expect(shot.status).toBe(200);
    const bad = await uploadAs(pl, Buffer.from('%!PS-Adobe not allowed'), 'x.ps');
    expect(bad.data.message || bad.data.error || JSON.stringify(bad.data)).toContain('ATTACHMENT_TYPE');
    const created = await rpc(pl, 'base_support_ticket_create', { departmentKey: 'payments', subject: 'سکه نرسید ' + k, body: 'خریدم ولی سکه نیامد', attachments: [shot.data.key], lang: 'fa' });
    expect(created.status).toBe(200);
    const id = created.data.id as string;
    const number = created.data.number as string;
    expect(number).toMatch(/^T-[A-Z2-9]{6}$/);

    await open(page, '#/support/tickets?status=waiting');
    await expect.poll(() => page.evaluate(() => (window as any).NakamaConsole.badges().tickets)).toBeGreaterThan(0);
    await expect(root(page).locator('[data-support-stats]')).toContainText('منتظر پاسخ');
    const row = rows(page).filter({ hasText: number });
    await expect(row).toContainText('سکه نرسید ' + k);
    await expect(row.locator('[data-ticket-status="open"]')).toHaveText('تازه');
    await row.click();
    await expect(page).toHaveURL(new RegExp('#/support/tickets/' + id));
    const tk = root(page).locator('[data-ticket]');
    await expect(tk.locator('[data-thread]')).toContainText('خریدم ولی سکه نیامد');
    const att = tk.locator('[data-attachment]').first();
    await expect(att).toContainText('screenshot.png');
    const href = await att.getAttribute('href');
    const file = await fetch(href!);
    expect(file.status).toBe(200);
    expect(file.headers.get('content-type')).toBe('image/png');
    expect((await fetch(href!.replace(/s=[0-9a-f]+/, 's=00'))).status).toBe(404);

    const composer = tk.locator('[data-composer]');
    await tk.locator('[data-internal-toggle] input, [data-internal-toggle]').first().click();
    await composer.fill('پرداخت در درگاه موفق بوده ' + k);
    await tk.locator('[data-send]').click();
    await expect(toasts(page).last()).toContainText('یادداشت ذخیره شد');
    await expect(tk.locator('[data-msg-kind="note"]')).toContainText('پرداخت در درگاه موفق بوده');
    await tk.locator('select.eco-canned').selectOption({ label: 'پیگیری ' + k });
    await expect(composer).toHaveValue(new RegExp('پرداختت را پیگیری کردیم ' + k));
    await tk.locator('[data-reply-status] [data-value="resolved"]').click();
    await tk.locator('[data-send]').click();
    await expect(toasts(page).last()).toContainText('بازیکن اعلان گرفت');
    await expect(tk.locator('[data-ticket-status]').first()).toHaveText('حل شده');
    await expect(tk.locator('[data-assignee]')).toHaveValue(/.+/);
    await expect(tk.locator('[data-sla]')).not.toContainText('هنوز نه');

    const mine = await rpc(pl, 'base_support_ticket_get', { id, lang: 'fa' });
    const bodies = mine.data.messages.map((m: any) => m.body).join('|');
    expect(bodies).toContain('پرداختت را پیگیری کردیم');
    expect(bodies).not.toContain('پرداخت در درگاه موفق بوده');
    expect(mine.data.status).toBe('resolved');
    expect(mine.data.canRate).toBe(true);
    await expect.poll(async () => {
      const inbox = await rpc(pl, 'publication_inbox', { lang: 'fa' });
      return (inbox.data.items || []).map((m: any) => m.title).join('|');
    }, { timeout: 20_000 }).toContain('تیکت شما حل شد');
    const rate = await rpc(pl, 'base_support_ticket_rate', { id, rating: 4, comment: 'ممنون ' + k });
    expect(rate.data.rating).toBe(4);

    await page.reload();
    await settle(page);
    await expect(tk.locator('[data-rating="4"]')).toContainText('ممنون ' + k);
    await tk.locator('[data-priority-pick] [data-value="urgent"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    await tk.locator('[data-move="closed"]').click();
    await expect(tk.locator('[data-ticket-status]').first()).toHaveText('بسته');
    const again = await rpc(pl, 'base_support_ticket_reply', { id, body: 'باز هم' });
    expect(JSON.stringify(again.data)).toContain('TICKET_CLOSED');
    await expect(tk.locator('[data-thread]')).toContainText('اولویت را فوری کرد');

    await open(page, '#/support/tickets?status=closed&priority=urgent');
    await expect(rows(page).filter({ hasText: number })).toHaveCount(1);
    const act = await base.get('/activity?userId=' + pl.id + '&kind=ticket');
    expect(act.items.map((x: any) => x.action)).toEqual(expect.arrayContaining(['created', 'staff_reply', 'rated', 'closed']));
  });

  test('a support agent creates a ticket for a player and the player gets it', async ({ page }) => {
    const pl = await newPlayer('staff');
    await open(page, '#/support/tickets');
    await root(page).locator('[data-ticket-new]').click();
    const dlg = page.locator('[data-form-dialog="ticket-new"]');
    await dlg.locator('button[aria-haspopup="listbox"]').first().click();
    await dlg.locator('.eco-pop:not([hidden]) input').fill(pl.username);
    await dlg.locator('.eco-pop:not([hidden]) [role="option"]').first().click();
    await dlg.locator('input[aria-label="موضوع"]').fill('پیگیری خرید ' + k);
    await dlg.locator('textarea[aria-label="پیام"]').fill('سلام، خریدت را بررسی کردیم.');
    await dlg.locator('[data-dialog-ok]').click();
    await expect(page).toHaveURL(/#\/support\/tickets\/[0-9a-f-]{36}/);
    await expect(root(page).locator('[data-ticket-status]').first()).toHaveText('منتظر بازیکن');
    const list = await rpc(pl, 'base_support_tickets', { lang: 'fa' });
    expect(list.data.items[0]).toMatchObject({ subject: 'پیگیری خرید ' + k, status: 'awaiting_player', unread: true });
  });
});

test.describe('support: access and search', () => {
  test('a support-only console user works tickets and live chat but cannot open economy or player management', async ({ spawn }) => {
    const user = await makeConsoleUser('e2esup' + k.slice(4, 12), { ACCOUNT_NOTES: { read: true, write: true } });
    try {
      const p = await spawn({ signedIn: false });
      await signIn(p, user);
      await p.goto(p.url().split('#')[0] + '#/support/tickets');
      await settle(p);
      const sections = await p.locator('.x-drawer [data-economy-nav]').evaluateAll((els) => els.map((e) => e.getAttribute('data-economy-nav')));
      expect(sections).toEqual(['support']);
      await p.goto(p.url().split('#')[0] + '#/project-settings/support-canned');
      await expect(psWindow(p)).toBeVisible();
      const ids = await psWindow(p).locator('.ps-list [role="tab"]').evaluateAll((els) => els.map((e) => e.getAttribute('data-ps-section')));
      expect(ids).toEqual(['support-departments', 'support-canned', 'support-settings']);
      const bearer = await bearerOf(p);
      expect((await consoleApi('GET', '/v2/console/base/support/tickets', undefined, bearer)).status).toBe(200);
      expect((await consoleApi('GET', '/v2/console/economy/products', undefined, bearer)).status).toBe(403);
      expect((await consoleApi('GET', '/v2/console/base/players', undefined, bearer)).status).toBe(403);
      expect((await consoleApi('GET', '/v2/console/base/activity', undefined, bearer)).status).toBe(403);
    } finally {
      await dropConsoleUser(user.user);
    }
  });

  test('every new page is in the page search', async ({ page }) => {
    await open(page, '#/support/tickets');
    for (const [q, url] of [['گفتگوی زنده', /#\/support\/live-chat/], ['فعالیت بازیکنان', /#\/moderation\/activity/], ['کدهای معرف', /#\/economy\/referrals/], ['پیکربندی بازی', /#\/publication\/remote-config/], ['تیکت‌ها', /#\/support\/tickets/]] as [string, RegExp][]) {
      await expect(page.locator('.sp-pal-input:visible')).toHaveCount(0);
      await page.keyboard.press('Control+k');
      await page.locator('.sp-pal-input:visible').fill(q);
      await expect(page.locator('.sp-pal-row').first()).toContainText(q);
      await page.keyboard.press('Enter');
      await expect(page).toHaveURL(url);
      await settle(page);
    }
  });
});

