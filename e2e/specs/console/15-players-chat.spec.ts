import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { API_URL, CONSOLE_URL, SERVER_KEY } from '../../lib/env';
import { base, consoleApi, deleteUser, devicePlayer, eco, must, pub, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { Chat, deviceAuth, history, notice, refreshSession } from '../../lib/chat';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, open, openSettings, panel, psWindow, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const room = 'e2e-' + k.replace(/_/g, '-').toLowerCase();
const made = { users: [] as string[], tags: [] as string[], products: [] as string[], prizes: [] as string[], words: [] as string[], titles: [] as string[] };
const chats: Chat[] = [];
let tag: any;
let A: Player;
let B: Player;
let C: Player;

test.describe.configure({ mode: 'serial' });

async function player(name: string): Promise<Player> {
  const p = await devicePlayer('e2e-' + name + '-' + k, (k + name).replace(/_/g, ''));
  made.users.push(p.id);
  await updateAccount(p, { lang_tag: 'fa' });
  return p;
}

async function chat(p: Player): Promise<Chat> {
  const c = await Chat.open(p);
  chats.push(c);
  await c.join(room);
  return c;
}

function dialog(page: Page, id: string): Locator {
  return page.locator(`[data-form-dialog="${id}"]`);
}

async function playerPage(page: Page, id: string) {
  await page.goto(CONSOLE_URL + '/#/players/' + id);
  await expect(page.locator('[data-player-mod] [data-player-status]')).toBeVisible({ timeout: 20_000 });
}

async function pickIn(scope: Locator, text: string) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

test.beforeAll(async () => {
  tag = await base.post('/tags', { key: 'e2e_' + k + '_vip', keyType: 'user', name: { fa: 'ویژهٔ ' + k, en: 'VIP ' + k }, description: {}, isPublic: true, metadata: {}, sort: 0 });
  made.tags.push(tag.id);
  A = await player('pa');
  B = await player('pb');
  C = await player('pc');
});

test.afterAll(async () => {
  for (const c of chats) c.close();
  for (const id of made.products) await eco.del('/products/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.words) await base.del('/chat-words/' + id);
  for (const id of made.titles) await base.del('/report-titles/' + id);
  for (const id of made.users) await deleteUser(id);
  for (const id of made.tags) await base.del('/tags/' + id);
  sql(`DELETE FROM base_chat_report WHERE channel_label LIKE 'e2e%'; DELETE FROM message WHERE stream_label LIKE 'e2e%';`);
});

test.describe('players: tags', () => {
  test('tags from the native player page, filter, bulk tagging, notification target and product audience', async ({ page }) => {
    expect(tag.isPublic).toBe(true);
    await playerPage(page, A.id);
    const card = page.locator('[data-player-mod]');
    await expect(card).toContainText('برچسب‌ها و محدودیت‌ها');
    await expect(card.locator('[data-player-status]')).toContainText('بدون محدودیت');
    await card.locator('[data-player-tags-edit]').click();
    const dlg = dialog(page, 'player-tags');
    await expect(dlg).toBeVisible();
    await pickIn(dlg, 'ویژهٔ ' + k);
    await dlg.locator('[data-dialog-ok]').click();
    await expect(toasts(page).filter({ hasText: 'برچسب‌ها ذخیره شد.' })).toBeVisible();
    await expect(page.locator('[data-player-mod] [data-player-tags]')).toContainText('ویژهٔ ' + k);
    const mine = await rpc(A, 'base_my_tags', { lang: 'en' });
    expect(mine.data.items.map((i: any) => i.name)).toEqual(['VIP ' + k]);
    expect((await rpc(B, 'base_my_tags', {})).data.items).toEqual([]);

    await go(page, '#/moderation/players?tagId=' + tag.id);
    await settle(page);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText(A.username);

    await go(page, '#/moderation/players');
    await search(page, k.replace(/_/g, ''));
    await expect(rows(page)).toHaveCount(3);
    await root(page).locator(`[data-select="${B.id}"]`).click();
    await root(page).locator(`[data-select="${C.id}"]`).click();
    const bar = root(page).locator('[data-bulk-bar]');
    await expect(bar).toContainText('۲ بازیکن انتخاب شده');
    await pickIn(bar, 'ویژهٔ ' + k);
    await bar.locator('[data-bulk="add"]').click();
    await expect(toasts(page).filter({ hasText: 'برچسب برای ۲ بازیکن ثبت شد.' })).toBeVisible();
    await expect(rows(page).filter({ hasText: 'ویژهٔ ' + k })).toHaveCount(3);
    expect((await base.get('/players?tagId=' + tag.id)).total).toBe(3);

    const preview = await pub.get(`/notifications/preview-target?mode=tags&tagIds=${tag.id}`);
    expect(preview.count).toBe(3);
    await go(page, '#/publication/notifications/new?tagId=' + tag.id);
    await expect(root(page).locator('[data-recipient-count="3"]')).toBeVisible();

    await go(page, '#/moderation/players?q=' + k.replace(/_/g, ''));
    await settle(page);
    await root(page).locator(`[data-select="${B.id}"]`).click();
    await root(page).locator(`[data-select="${C.id}"]`).click();
    await pickIn(root(page).locator('[data-bulk-bar]'), 'ویژهٔ ' + k);
    await root(page).locator('[data-bulk="remove"]').click();
    await expect(toasts(page).filter({ hasText: 'برچسب از ۲ بازیکن برداشته شد.' })).toBeVisible();
    expect((await base.get('/players?tagId=' + tag.id)).total).toBe(1);

    const prize = await eco.post('/prizes', { key: 'e2e' + k + 'p', name: { fa: 'جایزهٔ ' + k }, message: {}, tagIds: [], isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }], metadata: {} });
    made.prizes.push(prize.id);
    const product = await eco.post('/products', { key: 'e2e' + k + 'vip', name: { fa: 'بستهٔ ' + k }, description: {}, isActive: true, tagIds: [], audienceTagIds: [tag.id], paymentMethods: ['in_app'], coinPrice: { currencyKey: 'coin', amount: 1 }, discountPercent: 0, sku: {}, prizeId: prize.id, minLevel: 0, sort: 99, metadata: {} });
    made.products.push(product.id);
    const forA = await rpc(A, 'economy_catalog', {});
    expect(forA.data.products.find((p: any) => p.id === product.id)).toMatchObject({ exclusive: true });
    const forB = await rpc(B, 'economy_catalog', {});
    expect(forB.data.products.some((p: any) => p.id === product.id)).toBe(false);
    const buy = await rpc(B, 'economy_purchase_start', { productId: product.id, method: 'in_app' });
    expect(buy.data.message || buy.data).toContain('PRODUCT_NOT_FOUND');
    const general = await base.post('/tags', { key: 'e2e_' + k + '_gen', keyType: 'general', name: { fa: 'عمومی ' + k }, description: {}, metadata: {}, sort: 0 });
    made.tags.push(general.id);
    const wrong = await consoleApi('PUT', '/v2/console/economy/products/' + product.id, { ...product, audienceTagIds: [general.id] });
    expect(wrong.status).toBe(400);
    expect(wrong.data.error).toBe('TAG_KIND_MISMATCH');
    await go(page, '#/economy/products/' + product.id);
    await settlePanel(page);
    await expect(field(panel(page), 'فقط برای این بازیکنان')).toContainText('ویژهٔ ' + k);
  });
});

test.describe('players: suspension', () => {
  test('suspend from the player card: sign-in and refresh refused with the message and end time, history, lift early', async ({ page }) => {
    const auth = await fetch(API_URL + '/v2/account/authenticate/device?create=false', { method: 'POST', headers: { Authorization: 'Basic ' + Buffer.from(SERVER_KEY + ':').toString('base64'), 'Content-Type': 'application/json' }, body: JSON.stringify({ id: B.deviceId }) });
    const session = await auth.json();
    B.token = session.token;
    await playerPage(page, B.id);
    const card = page.locator('[data-player-mod]');
    await expect(card.locator('[data-restrict="suspend"]')).toBeVisible();
    await card.locator('[data-restrict="suspend"]').click();
    const dlg = dialog(page, 'restrict-suspend');
    await expect(dlg).toContainText('تعلیق موقت');
    await dlg.locator('[data-duration] button', { hasText: 'تاریخ دلخواه' }).click();
    await dlg.locator('[data-dialog-ok]').click();
    await expect(dlg).toContainText('زمان پایان را انتخاب کنید.');
    await dlg.locator('[data-duration] button', { hasText: '۳ روز' }).click();
    await fillTranslated(dlg, 'reason', { fa: 'e2e به خاطر آزار دیگران', en: 'e2e for bothering others' });
    await dlg.locator('textarea[aria-label="یادداشت داخلی"]').fill('e2e یادداشت');
    await dlg.locator('[data-dialog-ok]').click();
    await expect(toasts(page).filter({ hasText: 'تعلیق شد.' })).toBeVisible();
    await expect(card.locator('[data-status="suspended"]')).toContainText('معلق تا');

    const refused = await deviceAuth(B.deviceId);
    expect(refused.status).toBe(403);
    const n = notice(refused.data.message);
    expect(n).toMatchObject({ code: 'ACCOUNT_SUSPENDED', reason: 'e2e به خاطر آزار دیگران', lang: 'fa' });
    const hours = (new Date(n.until).getTime() - Date.now()) / 3_600_000;
    expect(hours).toBeGreaterThan(71);
    expect(hours).toBeLessThan(73);
    const refreshed = await refreshSession(session.refresh_token);
    expect(refreshed.status).toBe(403);
    expect(notice(refreshed.data.message).code).toBe('ACCOUNT_SUSPENDED');
    await expect.poll(async () => (await rpc(B, 'base_my_status', {})).status).toBe(401);
    expect((await base.get('/players?status=suspended&q=' + encodeURIComponent(B.username))).total).toBe(1);

    await go(page, '#/moderation/restrictions?userId=' + B.id);
    await settle(page);
    const row = rows(page).filter({ hasText: 'e2e به خاطر آزار دیگران' });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText('تعلیق');
    await expect(row).toContainText('فعال');
    await row.locator('[data-lift]').click();
    const lift = dialog(page, 'lift');
    await lift.locator('textarea').fill('e2e اشتباه شد');
    await lift.locator('[data-dialog-ok]').click();
    await expect(toasts(page).filter({ hasText: 'محدودیت برداشته شد.' })).toBeVisible();
    await expect(rows(page).filter({ hasText: 'e2e به خاطر آزار دیگران' })).toContainText('برداشته شده');
    const again = await deviceAuth(B.deviceId);
    expect(again.status).toBe(200);
    B.token = again.data.token;
    const hist = await base.get('/players/' + B.id);
    expect(hist.history[0]).toMatchObject({ kind: 'suspend', status: 'lifted', liftNote: 'e2e اشتباه شد', note: 'e2e یادداشت' });
  });
});

test.describe('chat: word filter', () => {
  test('Persian evasions are masked, block refuses, flag queues a report; the word list, try box and editor', async ({ page }) => {
    const a = await chat(A);
    const sent = await a.send('سلام، تو اح‌ـــمَق هستی');
    expect(sent.ok).toBe(true);
    const h1 = await history(A, a.channelId, 5);
    expect(h1[0].text).toBe('سلام، تو **** هستی');
    const blocked = await a.send('تو حرام‌زاده‌ای');
    expect(blocked).toEqual({ ok: false, message: 'CHAT_BLOCKED' });
    const flagged = await a.send('بیا تلگرام ' + k);
    expect(flagged.ok).toBe(true);
    await expect.poll(async () => (await base.get('/reports?source=filter&q=' + encodeURIComponent(k))).total).toBe(1);
    const flag = (await base.get('/reports?source=filter&q=' + encodeURIComponent(k))).items[0];
    expect(flag).toMatchObject({ channelLabel: room, status: 'open', senderId: A.id });
    expect(flag.matched.map((m: any) => m.word)).toEqual(['تلگرام']);

    await open(page, '#/moderation/chat-words');
    const tryBox = root(page).locator('[data-word-try]');
    await tryBox.fill('you are an i.d.i.o.t');
    await expect(root(page).locator('[data-word-shown]')).toHaveText('you are an *.*.*.*.*');
    await tryBox.fill('كثافت');
    await expect(root(page).locator('[data-word-shown]')).toHaveText('*****');
    await expect(root(page).locator('[data-word-normalized]')).toHaveText('کثافت');

    const word = 'e2e' + k.replace(/_/g, '').slice(-8) + 'bad';
    await go(page, '#/moderation/chat-words/new');
    await settlePanel(page);
    await field(panel(page), 'کلمه یا عبارت').locator('input').fill(word);
    await field(panel(page), 'چه شود؟').locator('select').selectOption('block');
    await savePanel(page);
    await search(page, word);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('رد پیام');
    const created = (await base.get('/chat-words?q=' + word)).items[0];
    made.words.push(created.id);
    expect((await a.send('hello ' + word)).ok).toBe(false);
    await tryBox.fill('hello ' + word.toUpperCase());
    await expect(root(page).locator('[data-word-result]')).toContainText('این پیام فرستاده نمی‌شود.');

    await rows(page).first().locator('td').nth(1).click();
    await settlePanel(page);
    await field(panel(page), 'چه شود؟').locator('select').selectOption('mask');
    await savePanel(page);
    expect((await a.send('hello ' + word + 's')).ok).toBe(true);
    expect((await history(A, a.channelId, 1))[0].text).toBe('hello ' + '*'.repeat(word.length + 1));

    await go(page, '#/moderation/chat-words/new');
    await settlePanel(page);
    await field(panel(page), 'کلمه یا عبارت').locator('input').fill('ﺍﺣﻤﻖ');
    await field(panel(page), 'چه شود؟').locator('select').selectOption('mask');
    await panel(page).locator('.x-form > div:last-child button.eco-btn').last().click();
    await expect(field(panel(page), 'کلمه یا عبارت')).toContainText('این مقدار قبلاً استفاده شده است.');
  });
});

test.describe('chat: reports', () => {
  test('a player reports with a reason; the review page shows the context and mutes the sender, who is refused with the end time', async ({ page }) => {
    const a = await chat(A);
    const b = await chat(B);
    expect((await b.send('e2e سلام به همه')).ok).toBe(true);
    const bad = await a.send('e2e پیام آزاردهنده ' + k);
    expect(bad.ok).toBe(true);
    expect((await b.send('e2e چرا این‌طوری می‌گی؟')).ok).toBe(true);
    if (!bad.ok) return;
    const titles = await rpc(B, 'base_report_titles', { lang: 'fa' });
    expect(titles.data.items.map((t: any) => t.key)).toEqual(expect.arrayContaining(['abuse', 'harassment', 'spam', 'other']));
    const rep = await rpc(B, 'base_report_message', { messageId: bad.id, titleKey: 'harassment', note: 'e2e اذیتم کرد' });
    expect(rep.status).toBe(200);
    expect(rep.data.duplicate).toBe(false);
    expect((await rpc(B, 'base_report_message', { messageId: bad.id, titleKey: 'abuse' })).data.duplicate).toBe(true);
    expect((await rpc(A, 'base_report_message', { messageId: bad.id, titleKey: 'abuse' })).data.message).toBe('CANNOT_REPORT_OWN');
    expect((await rpc(C, 'base_report_message', { messageId: bad.id, titleKey: 'nope' })).data.message).toBe('REPORT_TITLE_NOT_FOUND');

    await open(page, '#/moderation/reports');
    await expect(page.locator('.pl-side a[href^="#/moderation/reports"] .pl-badge')).toBeVisible();
    await search(page, k);
    const row = rows(page).filter({ hasText: 'e2e پیام آزاردهنده' });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText('آزار یا تهدید');
    await row.locator('td').first().click();
    await expect(page).toHaveURL(new RegExp('#/moderation/reports/' + rep.data.reportId));
    const detail = root(page).locator('[data-report]');
    await expect(detail.locator('[data-report-context] [data-reported]')).toContainText('e2e پیام آزاردهنده');
    await expect(detail.locator('[data-report-context]')).toContainText('e2e سلام به همه');
    await expect(detail.locator('[data-report-context]')).toContainText('e2e چرا این‌طوری می‌گی؟');
    await expect(detail).toContainText('e2e اذیتم کرد');
    await detail.locator('[data-report-actions] button[data-value="mute"]').click();
    const form = detail.locator('[data-restrict-form="mute"]');
    await form.locator('[data-duration] button', { hasText: '۱ ساعت' }).click();
    await fillTranslated(form, 'reason', { fa: 'e2e یک ساعت استراحت کن', en: 'e2e take an hour off' });
    await detail.locator('[data-report-submit]').click();
    await expect(toasts(page).filter({ hasText: 'فرستنده بی‌صدا شد.' })).toBeVisible();
    await expect(detail).toContainText('بررسی شده');

    const refused = await a.send('e2e باز هم من');
    expect(refused.ok).toBe(false);
    const n = notice((refused as any).message);
    expect(n).toMatchObject({ code: 'CHAT_MUTED', reason: 'e2e یک ساعت استراحت کن' });
    const mins = (new Date(n.until).getTime() - Date.now()) / 60_000;
    expect(mins).toBeGreaterThan(58);
    expect(mins).toBeLessThan(61);
    const status = await rpc(A, 'base_my_status', { lang: 'en' });
    expect(status.data.mute).toMatchObject({ code: 'CHAT_MUTED', reason: 'e2e take an hour off' });
    const report = await base.get('/reports/' + rep.data.reportId);
    expect(report).toMatchObject({ status: 'resolved', action: 'mute' });
    expect(report.restrictionId).toBeTruthy();
    await base.post('/restrictions/' + report.restrictionId + '/lift', { note: 'e2e' });
    expect((await a.send('e2e آزاد شدم')).ok).toBe(true);

    const second = await a.send('e2e پیام دوم ' + k);
    if (!second.ok) throw new Error('second message refused');
    const rep2 = await rpc(C, 'base_report_message', { messageId: second.id, titleKey: 'spam' });
    await go(page, '#/moderation/reports/' + rep2.data.reportId);
    await detail.locator('[data-report-actions] button[data-value="delete"]').click();
    await detail.locator('[data-report-submit]').click();
    await expect(toasts(page).filter({ hasText: 'پیام پاک شد.' })).toBeVisible();
    await expect(detail).toContainText('پیام پاک شده');
    expect((await history(A, a.channelId, 10)).some((m) => m.id === second.id && m.code === 0)).toBe(false);
    expect((await base.get('/reports/' + rep2.data.reportId)).messageDeleted).toBe(true);
  });
});

test.describe('chat: report reasons', () => {
  test('report reasons are translated in project settings and drive the list the game shows', async ({ page }) => {
    await openSettings(page, 'report-titles');
    await expect(rows(page).filter({ hasText: 'توهین یا فحاشی' })).toHaveCount(1);
    await go(page, '#/project-settings/report-titles/new');
    await settlePanel(page);
    await field(panel(page), 'کلید').locator('input').fill('e2e_' + k);
    await fillTranslated(panel(page), 'name', { fa: 'e2e کلاهبرداری', en: 'e2e Scam' });
    await savePanel(page);
    const created = (await base.get('/report-titles?q=e2e_' + k)).items[0];
    made.titles.push(created.id);
    const en = await rpc(C, 'base_report_titles', { lang: 'en' });
    expect(en.data.items.map((i: any) => i.name)).toContain('e2e Scam');
    await base.put('/report-titles/' + created.id, { ...created, isActive: false });
    const after = await rpc(C, 'base_report_titles', { lang: 'en' });
    expect(after.data.items.map((i: any) => i.name)).not.toContain('e2e Scam');
  });
});

test.describe('players and chat: access', () => {
  test('a console user with chat access sees reports and the word filter but not player management', async ({ spawn }) => {
    const user = { user: 'e2epm' + k.slice(4, 12), pass: 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!' };
    const add = await consoleApi('POST', '/v2/console/user', { username: user.user, email: user.user + '@example.com', acl: { CHANNEL_MESSAGE: { read: true, write: true } } });
    expect(add.status).toBeLessThan(300);
    try {
      const reset = await must(consoleApi('POST', '/v2/console/user/' + user.user + '/reset/password'));
      const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: reset.code || add.data?.token, password: user.pass }) });
      expect(pc.ok).toBe(true);
      const p = await spawn({ signedIn: false });
      await p.goto(CONSOLE_URL + '/#/login');
      await p.locator('input[name="username"]').fill(user.user);
      await p.locator('input[name="password"]').fill(user.pass);
      await p.keyboard.press('Enter');
      await p.locator('.x-drawer [data-economy-nav="moderation"]').first().waitFor({ state: 'attached', timeout: 20_000 });
      await p.goto(CONSOLE_URL + '/#/moderation/reports');
      await settle(p);
      const pages = await p.locator('.x-drawer [data-economy-nav="moderation"] a[data-economy-page]').evaluateAll((els) => els.map((e) => e.getAttribute('data-economy-page')));
      expect(pages).toEqual(['reports', 'chat-words']);
      await p.goto(CONSOLE_URL + '/#/project-settings/report-titles');
      await expect(psWindow(p)).toBeVisible();
      const ids = await psWindow(p).locator('.ps-list [role="tab"]').evaluateAll((els) => els.map((e) => e.getAttribute('data-ps-section')));
      expect(ids).toEqual(['report-titles']);
      const tok = await p.evaluate(() => JSON.parse(localStorage.getItem('SESSION_TOKENS') || '{}'));
      const bearer = 'Bearer ' + tok.accessToken;
      expect((await consoleApi('GET', '/v2/console/base/players', undefined, bearer)).status).toBe(403);
      expect((await consoleApi('GET', '/v2/console/base/chat-words', undefined, bearer)).status).toBe(200);
    } finally {
      await consoleApi('DELETE', '/v2/console/user?username=' + user.user);
    }
  });
});
