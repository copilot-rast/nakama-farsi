import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, deleteUser, devicePlayer, must, pub, rpc, uid, type Player } from '../../lib/api';
import { field, fillTranslated, go, open, panel, root, rows, savePanel, settle, settlePanel, toasts, translated } from '../../lib/console';
import { FIXTURES } from '../../lib/env';
import { openGame, persian, screen, toast } from '../../lib/game';

const k = uid();
const made = { faqs: [] as string[], news: [] as string[], categories: [] as string[], users: [] as string[] };

test.beforeAll(async () => {
  const providers = await pub.get('/providers');
  const live = providers.items.filter((p: any) => p.isActive && !p.sandbox);
  expect(live, 'every active SMS/email provider must be in sandbox before this spec sends anything').toEqual([]);
});

test.afterAll(async () => {
  for (const id of made.news) await pub.del('/news/' + id);
  for (const id of made.faqs) await pub.del('/faqs/' + id);
  for (const id of made.categories) await base.del('/categories/' + id);
  for (const id of made.users) await deleteUser(id);
});

async function newPlayer(tag: string, extra: { phone?: string; email?: string; lang?: string } = {}): Promise<Player> {
  const p = await devicePlayer('e2e-' + tag + '-' + k, (k + tag).replace(/_/g, ''));
  made.users.push(p.id);
  const body: Record<string, string> = {};
  if (extra.phone) body.metadata = JSON.stringify({ phone: extra.phone });
  if (extra.email) body.email = extra.email;
  if (extra.lang) body.lang_tag = extra.lang;
  if (Object.keys(body).length) await must(consoleApi('POST', '/v2/console/account/' + p.id, body));
  return p;
}

function tile(page: import('@playwright/test').Page, label: string) {
  return root(page).locator('.eco-tile-grid > div').filter({ has: page.locator('div', { hasText: new RegExp('^' + label + '$') }) }).locator('.text-2xl');
}

async function pickUser(page: import('@playwright/test').Page, username: string) {
  const f = field(root(page), /^کاربران$/);
  await f.locator('button[aria-haspopup="listbox"]').click();
  await f.locator('.eco-pop input').fill(username);
  await f.locator('[role="option"]', { hasText: username }).click();
  await expect(f.locator('.eco-tag', { hasText: username })).toBeVisible();
}

test.describe('publication', () => {
  test('faq with a markdown answer shows up in the game', async ({ page, spawn }) => {
    const cat = await base.post('/categories', { key: k + 'f', keyType: 'faq', name: { fa: 'راهنمای ' + k, en: 'Help ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    made.categories.push(cat.id);
    await open(page, '#/publication/faqs/new');
    await settlePanel(page);
    const p = panel(page);
    await fillTranslated(p, 'question', { fa: 'چطور سکه بگیرم؟ ' + k, en: 'How do I get coins? ' + k });
    await fillTranslated(p, 'answer', { fa: 'هر دور **بازی** کن یا [فروشگاه](https://example.com/shop) را ببین.\n\n- یک\n- دو' });
    const answer = translated(p, 'answer');
    await answer.locator('[role="tab"][data-lang="fa"]').click();
    await answer.locator('[role="tabpanel"][data-lang="fa"] button[data-value="preview"]').click();
    const preview = answer.locator('[role="tabpanel"][data-lang="fa"] .eco-md');
    await expect(preview.locator('strong')).toHaveText('بازی');
    await expect(preview.locator('a[href="https://example.com/shop"]')).toHaveText('فروشگاه');
    await expect(preview.locator('li')).toHaveCount(2);
    await answer.locator('[role="tabpanel"][data-lang="fa"] button[data-value="write"]').click();
    const catField = field(p, /^دسته‌بندی$/);
    await catField.locator('button[aria-haspopup="listbox"]').click();
    await catField.locator('.eco-pop input').fill(k + 'f');
    await catField.locator('[role="option"]').first().click();
    await savePanel(page);
    const list = await pub.get('/faqs?q=' + encodeURIComponent(k) + '&limit=5');
    expect(list.items).toHaveLength(1);
    made.faqs.push(list.items[0].id);
    expect(list.items[0].categoryId).toBe(cat.id);
    await go(page, '#/publication/faqs?q=' + k);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('راهنمای ' + k);

    const game = await spawn({ signedIn: false });
    made.users.push((await openGame(game, 'e2e-faq-' + k)).id);
    await game.locator('[data-go="faq"]').click();
    await screen(game, 'faq');
    await game.locator('#faq-q').fill(k);
    const item = game.locator('#faq-list details.faq', { hasText: k });
    await expect(item).toHaveCount(1);
    await expect(item.locator('summary')).toContainText('چطور سکه بگیرم؟');
    await item.locator('summary').click();
    await expect(item.locator('.rich strong')).toHaveText('بازی');
    await expect(item.locator('.rich a')).toHaveAttribute('target', '_blank');
    await expect(game.locator('#faq-tabs [data-cat="' + k + 'f"]')).toContainText('راهنمای ' + k);
  });

  test('news with cover, markdown body and actions; comment moderation round trip with the game', async ({ page, spawn }) => {
    await open(page, '#/publication/news/new');
    await settlePanel(page);
    const p = panel(page);
    await fillTranslated(p, 'title', { fa: 'خبر تازه ' + k, en: 'Fresh news ' + k });
    await fillTranslated(p, 'summary', { fa: 'خلاصهٔ خبر' });
    await fillTranslated(p, 'body', { fa: '## رویداد\n\nامروز **دو برابر** سکه بگیر.' });
    const body = translated(p, 'body');
    await body.locator('[role="tabpanel"][data-lang="fa"] button[data-value="preview"]').click();
    await expect(body.locator('[role="tabpanel"][data-lang="fa"] .eco-md h2')).toHaveText('رویداد');
    await expect(body.locator('[role="tabpanel"][data-lang="fa"] .eco-md strong')).toHaveText('دو برابر');
    const cover = field(p, /^تصویر کاور$/);
    const upload = page.waitForResponse((r) => r.url().includes('/v2/console/economy/uploads') && r.request().method() === 'POST');
    await cover.locator('input[type="file"]').setInputFiles(path.join(FIXTURES, 'logo.svg'));
    expect((await upload).ok()).toBe(true);
    await expect(cover.locator('input[aria-label]').first()).toHaveValue(/\/v2\/economy\/files\/.+\.svg$/);
    await expect(cover.locator('img')).toBeVisible();
    await p.getByRole('button', { name: 'افزودن دکمه' }).click();
    await p.getByRole('button', { name: 'افزودن دکمه' }).click();
    const acts = p.locator('[data-action-row]');
    await expect(acts).toHaveCount(2);
    await acts.nth(0).locator('input[aria-label="مقصد"]').fill('https://example.com/e2e');
    await fillTranslated(acts.nth(0), '', { fa: 'سایت' });
    await acts.nth(1).locator('select[aria-label="نوع"]').selectOption('goto');
    await acts.nth(1).locator('input[aria-label="مقصد"]').fill('shop');
    await fillTranslated(acts.nth(1), '', { fa: 'برو به فروشگاه' });
    await savePanel(page);
    const list = await pub.get('/news?q=' + encodeURIComponent(k) + '&limit=5');
    expect(list.items).toHaveLength(1);
    const news = list.items[0];
    made.news.push(news.id);
    expect(news.actions).toEqual([
      { type: 'url', value: 'https://example.com/e2e', label: { fa: 'سایت' } },
      { type: 'goto', value: 'shop', label: { fa: 'برو به فروشگاه' } },
    ]);
    expect(news.allowComment).toBe(true);
    expect(news.freeComment).toBe(false);
    await go(page, '#/publication/news?q=' + k);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first().locator('img')).toHaveAttribute('src', news.coverUrl);

    const game = await spawn({ signedIn: false });
    const me = await openGame(game, 'e2e-news-' + k);
    made.users.push(me.id);
    await expect(game.locator('#news-dot')).toBeVisible();
    await game.locator('#menu-news').click();
    await screen(game, 'news');
    const card = game.locator(`[data-news="${news.id}"]`);
    await expect(card).toContainText('خبر تازه ' + k);
    await expect(card.locator('.chip-new')).toBeVisible();
    await expect(card.locator('img.news-cover')).toHaveAttribute('src', news.coverUrl);
    await card.click();
    await screen(game, 'article');
    await expect(game.locator('#article h1')).toHaveText('خبر تازه ' + k);
    await expect(game.locator('#article .rich h2')).toHaveText('رویداد');
    await expect(game.locator('#article [data-action="url"]')).toContainText('سایت');
    const msg = 'نظر آزمایشی ' + k;
    await game.locator('#comment-text').fill(msg);
    await game.locator('#comment-form button[type="submit"]').click();
    await expect(toast(game)).toContainText('بعد از تأیید');
    const mine = game.locator('.comment', { hasText: msg });
    await expect(mine).toHaveClass(/pending/);
    await expect(mine).toContainText('در انتظار تأیید');

    const other = await newPlayer('other');
    let seen = await rpc(other, 'publication_news_comments', { newsId: news.id });
    expect(seen.data.items.map((c: any) => c.message)).not.toContain(msg);

    await open(page, '#/publication/comments?status=pending&newsId=' + news.id);
    const pendingTotal = (await pub.get('/comments?status=pending&limit=1')).total;
    await expect(page.locator('.x-drawer a[data-economy-page="comments"] .eco-nav-badge')).toHaveText(persian(pendingTotal));
    const crow = rows(page).filter({ hasText: msg });
    await expect(crow).toHaveCount(1);
    await expect(crow).toContainText('خبر تازه ' + k);
    await expect(crow).toContainText('در انتظار تأیید');
    await crow.getByRole('button', { name: 'تأیید' }).click();
    await expect(toasts(page).filter({ hasText: 'کامنت تأیید شد.' })).toBeVisible();
    await expect(rows(page).filter({ hasText: msg })).toHaveCount(0);
    seen = await rpc(other, 'publication_news_comments', { newsId: news.id });
    expect(seen.data.items.find((c: any) => c.message === msg)).toMatchObject({ isAccepted: true, mine: false });

    await game.locator('#screen-article [data-go="news"]').click();
    await game.locator(`[data-news="${news.id}"]`).click();
    const accepted = game.locator('.comment', { hasText: msg });
    await expect(accepted).not.toHaveClass(/pending/);
    await expect(accepted.locator('.link-btn', { hasText: 'پاسخ' })).toBeVisible();
    await game.locator('#article [data-action="goto"]').click();
    await screen(game, 'shop');

    await go(page, '#/publication/comments?status=accepted&newsId=' + news.id);
    const arow = rows(page).filter({ hasText: msg });
    await expect(arow.locator('button', { hasText: 'رد کردن' })).toHaveAttribute('title', 'کامنت از دید بازیکنان پنهان می‌شود.');
    await arow.locator('button', { hasText: 'رد کردن' }).click();
    await expect(toasts(page).filter({ hasText: 'کامنت رد شد.' })).toBeVisible();
    seen = await rpc(other, 'publication_news_comments', { newsId: news.id });
    expect(seen.data.items.map((c: any) => c.message)).not.toContain(msg);
    const own = await rpc({ ...other, token: me.token }, 'publication_news_comments', { newsId: news.id });
    expect(own.data.items.find((c: any) => c.message === msg)).toMatchObject({ isAccepted: false, mine: true });

    await go(page, '#/publication/comments?status=pending&newsId=' + news.id);
    const drow = rows(page).filter({ hasText: msg });
    await drow.getByRole('button', { name: 'حذف' }).click();
    await drow.getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(toasts(page).filter({ hasText: 'حذف شد.' })).toBeVisible();
    await expect(rows(page).filter({ hasText: msg })).toHaveCount(0);
  });

  test('in-app notification reaches the open game live: toast, bell count, inbox', async ({ page, spawn }) => {
    const game = await spawn({ signedIn: false });
    const me = await openGame(game, 'e2e-inbox-' + k);
    made.users.push(me.id);
    await expect(game.locator('body')).toHaveAttribute('data-socket', 'open');
    await expect(game.locator('#inbox-badge')).toBeHidden();

    await open(page, '#/publication/notifications/new');
    await settle(page);
    await expect(root(page).locator('button[role="radio"][data-value="in_app"]')).toHaveAttribute('aria-checked', 'true');
    const title = 'پیام زنده ' + k;
    await fillTranslated(field(root(page), /^عنوان$/), '', { fa: title });
    await fillTranslated(field(root(page), /^متن$/), '', { fa: 'سلام **قهرمان**!' });
    await pickUser(page, me.username);
    await expect(root(page).locator('[data-recipient-count="1"]')).toBeVisible();
    await root(page).locator('.eco-compose-aside').getByRole('button', { name: 'ارسال اعلان' }).click();
    await expect(page).toHaveURL(/#\/publication\/notifications\/[0-9a-f-]{36}/);
    await expect(toasts(page).filter({ hasText: 'اعلان در صف ارسال قرار گرفت.' })).toBeVisible();

    await expect(game.locator('#note-toast')).toBeVisible({ timeout: 20_000 });
    await expect(game.locator('#note-toast')).toContainText(title);
    await expect(game.locator('#inbox-badge')).toBeVisible();
    await expect(game.locator('#inbox-badge')).toHaveText('۱');

    await expect(root(page)).toContainText('انجام شد', { timeout: 20_000 });
    await expect(tile(page, 'ارسال‌شده')).toHaveText('۱');
    await expect(tile(page, 'گیرندگان')).toHaveText('۱');

    await game.locator('#btn-inbox').click();
    await screen(game, 'inbox');
    const item = game.locator('.inbox-item.unread', { hasText: title });
    await expect(item).toBeVisible();
    await expect(item.locator('.rich strong')).toHaveText('قهرمان');
    await game.locator('#btn-seen-all').click();
    await expect(game.locator('#inbox-badge')).toBeHidden();
    await expect(game.locator('.inbox-item.unread', { hasText: title })).toHaveCount(0);
    await page.reload();
    await settle(page);
    await expect(tile(page, 'دیده‌شده')).toHaveText('۱');
  });

  for (const ch of [{ key: 'sms', label: 'پیامک' }, { key: 'email', label: 'ایمیل' }] as const) {
    test(`${ch.key} sandbox to one user with an address and one without`, async ({ page }) => {
      const withAddr = await newPlayer(ch.key + 'a', { phone: '09120000000', email: ch.key + '.' + k + '@example.com', lang: 'en' });
      const without = await newPlayer(ch.key + 'b');
      await open(page, '#/publication/notifications/new');
      await settle(page);
      await root(page).locator(`button[role="radio"][data-value="${ch.key}"]`).click();
      await expect(root(page).locator('.eco-sandbox-note')).toContainText('حالت آزمایشی');
      await expect(root(page).locator('.eco-notice--error')).toHaveCount(0);
      const faText = 'متن آزمایشی ' + k;
      const enText = 'Sandbox text ' + k;
      if (ch.key === 'email') await fillTranslated(field(root(page), /^عنوان$/), '', { fa: 'موضوع ' + k, en: 'Subject ' + k });
      else await expect(field(root(page), /^عنوان$/)).toBeHidden();
      const bodyField = root(page).locator('.bg-white').filter({ has: page.locator('h2', { hasText: /^متن پیام$/ }) }).locator('.eco-field').filter({ has: page.locator('[role="tablist"]') }).last();
      await fillTranslated(bodyField, '', { fa: faText, en: enText });
      await pickUser(page, withAddr.username);
      await pickUser(page, without.username);
      await expect(root(page).locator('[data-recipient-count="2"]')).toBeVisible();
      await expect(root(page).locator('.eco-compose-aside')).toContainText(ch.key === 'sms' ? '۱ نفر شماره دارند' : '۱ نفر ایمیل دارند');
      await expect(root(page).locator('.eco-compose-aside')).toContainText('۱ نفر نشانی ندارند و کنار گذاشته می‌شوند.');
      await root(page).locator('.eco-compose-aside').getByRole('button', { name: 'ارسال اعلان' }).click();
      await expect(page).toHaveURL(/#\/publication\/notifications\/[0-9a-f-]{36}/);
      await expect(root(page)).toContainText('انجام شد', { timeout: 30_000 });
      await expect(tile(page, 'گیرندگان')).toHaveText('۲');
      await expect(tile(page, 'ارسال‌شده')).toHaveText('۱');
      await expect(tile(page, 'بدون نشانی')).toHaveText('۱');
      await expect(tile(page, 'ناموفق')).toHaveText('۰');
      const sentRow = rows(page).filter({ hasText: withAddr.username });
      await expect(sentRow).toContainText('ارسال‌شده');
      await expect(sentRow).toContainText('آزمایشی');
      await expect(sentRow).toContainText(ch.key === 'sms' ? '+989120000000' : ch.key + '.' + k + '@example.com');
      await expect(sentRow.locator('[data-recipient-lang="en"]')).toBeVisible();
      await expect(rows(page).filter({ hasText: without.username })).toContainText('بدون نشانی');
      const id = page.url().split('/').pop()!.split('?')[0];
      const detail = await pub.get('/notifications/' + id);
      const r = detail.recipients.items.find((x: any) => x.userId === withAddr.id);
      expect(r).toMatchObject({ status: 'sent', error: 'sandbox', lang: 'en' });
      expect(r.message).toContain(enText);
      expect(detail.recipients.items.find((x: any) => x.userId === without.id).status).toBe('skipped');
    });
  }

  test('provider test-send answers from the sandbox for SMS and email', async ({ page }) => {
    for (const [tab, kind, to] of [['sms', 'kavenegar', '09120000000'], ['email', 'smtp', 'e2e@example.com']]) {
      await open(page, '#/project-settings/' + tab);
      const card = root(page).locator(`[data-provider="${kind}"]`);
      await expect(card.locator('[data-sandbox-box]')).toHaveClass(/eco-sandbox-note/);
      await card.locator('details.eco-test > summary').click();
      await card.locator('details.eco-test input[aria-label="گیرنده"]').fill(to);
      await card.locator('details.eco-test button', { hasText: 'ارسال آزمایشی' }).click();
      await expect(card.locator('[data-test-result="sandbox"]')).toBeVisible();
      await expect(card.locator('[data-test-result="sandbox"]')).toContainText('حالت آزمایشی: چیزی ارسال نشد.');
    }
  });
});
