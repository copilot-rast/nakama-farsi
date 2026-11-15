import crypto from 'node:crypto';
import { expect, test } from '../../lib/fixtures';
import { API_URL } from '../../lib/env';
import { consoleApi, deleteUser, devicePlayer, pub, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, root, rows, savePanel, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const srcKey = ('e2e-' + k).replace(/_/g, '-').slice(0, 38);
const made = { users: [] as string[], sources: [] as string[] };
let P: Player;
let savedSettings = '';

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

async function ingest(key: string, body: string, headers: Record<string, string>) {
  const r = await fetch(API_URL + '/v2/publication/news/ingest/' + key, { method: 'POST', headers: { 'Content-Type': 'application/json', ...headers }, body });
  const text = await r.text();
  let data: any = text;
  try { data = JSON.parse(text); } catch { data = text; }
  return { status: r.status, data };
}

function sign(secret: string, ts: string, body: string) {
  return crypto.createHmac('sha256', secret).update(ts + '.' + body).digest('hex');
}

async function newsByTitle(title: string) {
  const res = await pub.get('/news?q=' + encodeURIComponent(title) + '&limit=50');
  return (res.items || []).find((x: any) => (x.title || {}).fa === title || (x.title || {}).en === title);
}

test.beforeAll(async () => {
  savedSettings = sql(`SELECT row_to_json(s)::text FROM publication_news_settings s WHERE id=1`);
  P = await devicePlayer('e2e-news-' + k, ('nw' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
  sql(`UPDATE users SET create_time = now() - interval '30 days' WHERE id=${q(P.id)}`);
});

test.afterAll(async () => {
  sql(`DELETE FROM publication_notification WHERE news_id IN (SELECT id FROM publication_news WHERE title::text LIKE '%e2e%' OR source LIKE 'webhook:e2e%')`);
  sql(`DELETE FROM publication_news WHERE title::text LIKE '%e2e%' OR source LIKE 'webhook:e2e%'`);
  sql(`DELETE FROM publication_news_source WHERE key LIKE 'e2e%'`);
  sql(`DELETE FROM publication_news_reaction_type WHERE key LIKE 'e2e%'`);
  if (savedSettings) sql(`UPDATE publication_news_settings s SET (comments_per_minute, reactions_per_minute, unread_scope, show_expired, fill_default_lang, realtime_on_publish, notify_category_key, log_retention_days, log_payload_bytes, slug_style, public_enabled, public_base_url, public_path, rss_enabled, public_cache_seconds, site_title, updated_by, updated_at) = (SELECT r.comments_per_minute, r.reactions_per_minute, r.unread_scope, r.show_expired, r.fill_default_lang, r.realtime_on_publish, r.notify_category_key, r.log_retention_days, r.log_payload_bytes, r.slug_style, r.public_enabled, r.public_base_url, r.public_path, r.rss_enabled, r.public_cache_seconds, r.site_title, r.updated_by, r.updated_at FROM json_populate_record(NULL::publication_news_settings, ${q(savedSettings)}::json) r) WHERE s.id=1`);
  for (const id of made.users) await deleteUser(id);
});

test.describe('news: console', () => {
  test('the news hub has articles, review, sources, log, reactions and settings tabs', async ({ page }) => {
    await open(page, '#/publication/news');
    await expect(root(page).locator('[data-economy-tabs="news"] [data-economy-tab]')).toHaveText(['خبرها', 'صف بررسی', 'منبع‌ها', 'گزارش دریافت', 'واکنش‌ها', 'تنظیمات']);
    await open(page, '#/publication/news/reactions');
    for (const key of ['like', 'love', 'haha', 'wow', 'sad']) await expect(rows(page).filter({ hasText: key })).toHaveCount(1);
  });

  test('rich editor: format, table, task list and SEO preview; the HTML is saved, reloads and reaches the game', async ({ page }) => {
    const title = 'خبر غنی e2e ' + k.slice(-4);
    await open(page, '#/publication/news/new');
    await settlePanel(page);
    const p = panel(page);
    await fillTranslated(p, 'title', { fa: title });
    const fa = p.locator('[data-translated="body"] [role="tabpanel"][data-lang="fa"]');
    const ed = fa.locator('.eco-rt-body');
    await expect(ed).toBeVisible();
    await expect(ed).toHaveAttribute('dir', 'rtl');
    await ed.click();
    await page.keyboard.type('سلام ');
    await fa.locator('[data-rt="bold"]').click();
    await page.keyboard.type('پررنگ');
    await fa.locator('[data-rt="bold"]').click();
    await page.keyboard.press('Enter');
    await fa.locator('[data-rt="h2"]').click();
    await page.keyboard.type('سرتیتر');
    await page.keyboard.press('Enter');
    await fa.locator('[data-rt="taskList"]').click();
    await page.keyboard.type('کار اول');
    await page.keyboard.press('Enter');
    await page.keyboard.press('Enter');
    await fa.locator('[data-rt="alignCenter"]').click();
    await page.keyboard.type('وسط');
    await page.keyboard.press('Enter');
    await fa.locator('[data-rt="table"]').click();
    await expect(fa.locator('[data-rt="rowAfter"]')).toBeVisible();
    await page.keyboard.type('خانه');
    await fa.locator('[data-rt="rowAfter"]').click();
    await expect(ed.locator('table tr')).toHaveCount(4);
    await fa.locator('[data-rt="undo"]').click();
    await expect(ed.locator('table tr')).toHaveCount(3);
    await expect(ed.locator('strong')).toHaveText('پررنگ');
    await fillTranslated(p, 'metaTitle', { fa: 'عنوان جست‌وجو ' + k.slice(-4) });
    await expect(p.locator('[data-news-serp] [data-serp-title]')).toHaveText('عنوان جست‌وجو ' + k.slice(-4));
    await field(p, 'کلیدواژه‌ها').locator('input').fill('یلدا، انار');
    await savePanel(page);
    const made1 = await newsByTitle(title);
    expect(made1.bodyFormat).toBe('html');
    const html = made1.body.fa as string;
    expect(html).toContain('<strong>پررنگ</strong>');
    expect(html).toContain('<h2>سرتیتر</h2>');
    expect(html).toContain('data-type="taskList"');
    expect(html).toContain('style="text-align: center"');
    expect(html).toContain('<table>');
    expect(html).not.toContain('class=');
    expect(made1.keywords).toEqual(['یلدا', 'انار']);
    expect(made1.slug).toBe(title.replace(/ /g, '-').toLowerCase());
    await open(page, '#/publication/news/' + made1.id);
    await settlePanel(page);
    const ed2 = panel(page).locator('[data-translated="body"] [role="tabpanel"][data-lang="fa"] .eco-rt-body');
    await expect(ed2.locator('table td, table th').first()).toBeVisible();
    await expect(ed2.locator('strong')).toHaveText('پررنگ');
    await expect(ed2.locator('ul[data-type="taskList"] > li')).toHaveCount(1);
    const got = await rpc(P, 'publication_news_get', { id: made1.id });
    expect(got.status).toBe(200);
    expect(got.data.bodyHtml).toContain('<table>');
    expect(got.data.seo.metaTitle).toBe('عنوان جست‌وجو ' + k.slice(-4));
    expect(got.data.seo.metaDescription).toContain('سلام پررنگ');
  });

  test('status column and filter; the review queue approves and rejects in bulk', async ({ page }) => {
    const ids: string[] = [];
    for (const n of ['۱', '۲', '۳']) {
      const r = await pub.post('/news', { title: { fa: 'در صف e2e ' + n + ' ' + k.slice(-4) }, bodyFormat: 'html', body: { fa: '<p>x</p>' }, status: 'pending_review' });
      ids.push(r.id);
    }
    await open(page, '#/publication/news?status=pending_review&q=' + encodeURIComponent(k.slice(-4)));
    await expect(rows(page)).toHaveCount(3);
    await expect(rows(page).first().locator('[data-news-status="pending_review"]')).toHaveText('در انتظار بررسی');
    await open(page, '#/publication/news/review');
    const bar = root(page).locator('[data-review-bar]');
    await root(page).locator(`[data-review-pick="${ids[0]}"]`).check();
    await root(page).locator(`[data-review-pick="${ids[1]}"]`).check();
    await expect(bar.locator('[data-review-count]')).toHaveText('۲ خبر انتخاب شده');
    await bar.locator('[data-review="approve"]').click();
    await expect(toasts(page).last()).toContainText('۲ خبر منتشر شد');
    await expect(root(page).locator(`[data-review-pick="${ids[0]}"]`)).toHaveCount(0);
    await root(page).locator(`[data-review-pick="${ids[2]}"]`).check();
    await bar.locator('[data-review-reason]').fill('تکراری');
    await bar.locator('[data-review="reject"]').click();
    await expect(toasts(page).last()).toContainText('۱ خبر رد شد');
    expect(sql(`SELECT status || ':' || reject_reason FROM publication_news WHERE id=${q(ids[2])}`)).toBe('rejected:تکراری');
    expect(sql(`SELECT count(*) FROM publication_news WHERE id IN (${q(ids[0])},${q(ids[1])}) AND status='published' AND approved_by <> ''`)).toBe('2');
  });

  test('a webhook source: token and HMAC auth, dedupe, review vs auto-publish, test payload, log and notify once', async ({ page }) => {
    await open(page, '#/publication/news/sources/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(srcKey);
    await field(p, 'مسیر فهرست خبرها').locator('input').fill('items');
    await field(p, 'شناسهٔ بیرونی').locator('input').fill('guid');
    await field(p, 'اعلان').locator('select').selectOption('in_app');
    await expect(p.locator('[data-src-endpoint] code')).toHaveText('/v2/publication/news/ingest/' + srcKey);
    await savePanel(page);
    const list = await pub.get('/news-sources?q=' + srcKey);
    const src = list.items[0];
    made.sources.push(src.id);
    expect(src.publishMode).toBe('review');
    await open(page, '#/publication/news/sources/' + src.id);
    await settlePanel(page);
    await panel(page).locator('[data-src-reveal]').click();
    const secret = (await panel(page).locator('[data-src-secret-value]').textContent())!.trim();
    expect(secret.length).toBeGreaterThan(30);
    await panel(page).locator('[data-src-test-payload]').fill(JSON.stringify({ items: [{ guid: 'pv', title: 'پیش‌نمایش e2e', body: '<p>Hi <script>x</script><b>there</b></p>' }] }));
    await panel(page).locator('[data-src-test="preview"]').click();
    await expect(panel(page).locator('[data-src-test-out] [data-src-result="created"]')).toBeVisible();
    await expect(panel(page).locator('[data-src-test-out] .eco-md strong')).toHaveText('there');
    expect(sql(`SELECT count(*) FROM publication_news WHERE external_id='pv'`)).toBe('0');

    const t1 = 'وب‌هوک e2e ' + k.slice(-4);
    const body = JSON.stringify({ items: [{ guid: 'g1-' + k, title: t1, body: '<p>متن <b>خبر</b></p><iframe src="x"></iframe>', lang: 'fa-IR' }] });
    expect((await ingest(srcKey, body, { 'X-News-Token': 'nope' })).status).toBe(401);
    const ok = await ingest(srcKey, body, { 'X-News-Token': secret });
    expect(ok.status).toBe(200);
    expect(ok.data.created).toBe(1);
    const newsId = ok.data.results[0].newsId;
    const again = await ingest(srcKey, body, { 'X-News-Token': secret });
    expect(again.data.unchanged).toBe(1);
    expect(sql(`SELECT status || ':' || lang FROM publication_news WHERE id=${q(newsId)}`)).toBe('pending_review:fa');
    expect(sql(`SELECT body->>'fa' FROM publication_news WHERE id=${q(newsId)}`)).toBe('<p>متن <strong>خبر</strong></p>');
    await page.waitForTimeout(6000);
    expect(sql(`SELECT count(*) FROM publication_notification WHERE news_id=${q(newsId)}`)).toBe('0');
    await pub.post('/news-review', { ids: [newsId], action: 'approve' });
    await expect.poll(() => sql(`SELECT count(*) FROM publication_notification WHERE news_id=${q(newsId)} AND kind='news'`), { timeout: 20_000 }).toBe('1');
    await page.waitForTimeout(6000);
    expect(sql(`SELECT count(*) FROM publication_notification WHERE news_id=${q(newsId)}`)).toBe('1');
    await expect.poll(async () => (await rpc(P, 'publication_inbox', {})).data.items.some((x: any) => x.title === t1), { timeout: 20_000 }).toBe(true);

    await pub.put('/news-sources/' + src.id, { ...src, authMode: 'hmac', publishMode: 'auto', mapping: { ...src.mapping } });
    const t2 = 'امضا e2e ' + k.slice(-4);
    const b2 = JSON.stringify({ items: [{ guid: 'g2-' + k, title: t2, body: '**md** body' }] });
    const ts = String(Math.floor(Date.now() / 1000));
    expect((await ingest(srcKey, b2, { 'X-News-Timestamp': ts, 'X-News-Signature': 'sha256=' + sign('wrong-secret-0000000', ts, b2) })).status).toBe(401);
    const signed = await ingest(srcKey, b2, { 'X-News-Timestamp': ts, 'X-News-Signature': 'sha256=' + sign(secret, ts, b2) });
    expect(signed.status).toBe(200);
    const replay = await ingest(srcKey, b2, { 'X-News-Timestamp': ts, 'X-News-Signature': 'sha256=' + sign(secret, ts, b2) });
    expect(replay.status).toBe(409);
    expect(replay.data.error).toBe('REPLAYED');
    const id2 = signed.data.results[0].newsId;
    expect(sql(`SELECT status FROM publication_news WHERE id=${q(id2)}`)).toBe('published');
    await expect.poll(() => sql(`SELECT count(*) FROM publication_notification WHERE news_id=${q(id2)}`), { timeout: 20_000 }).toBe('1');

    await open(page, '#/publication/news/log?sourceId=' + src.id);
    await expect(rows(page).first()).toBeVisible();
    await expect(rows(page).locator('[data-ingest-status="rejected"]').first()).toBeVisible();
    await rows(page).filter({ has: page.locator('[data-ingest-status="created"]') }).last().click();
    await expect(page.locator('[data-ingest-payload]')).toContainText('g1-' + k);
  });

  test('reactions and tags over RPC, unread count and seen', async () => {
    const news = await pub.post('/news', { title: { fa: 'واکنش e2e ' + k.slice(-4) }, bodyFormat: 'html', body: { fa: '<p>x</p>' }, status: 'published' });
    const before = (await rpc(P, 'publication_news_unread', {})).data.unread;
    expect(before).toBeGreaterThan(0);
    const r1 = await rpc(P, 'publication_news_react', { newsId: news.id, reaction: 'love' });
    expect(r1.status).toBe(200);
    expect(r1.data.myReaction).toBe('love');
    const r2 = await rpc(P, 'publication_news_react', { newsId: news.id, reaction: '' });
    expect(r2.data.reactionCount).toBe(0);
    const list = await rpc(P, 'publication_news_list', { limit: 50 });
    expect(list.data.reactionTypes.length).toBeGreaterThanOrEqual(5);
    await rpc(P, 'publication_news_seen', { all: true });
    expect((await rpc(P, 'publication_news_unread', {})).data.unread).toBe(0);
  });

  test('settings: slug style and the public pages switch; sitemap only lists published, indexable news', async ({ page }) => {
    const r1 = await pub.post('/news', { title: { fa: 'نامک e2e ' + k.slice(-4) }, bodyFormat: 'html', body: { fa: '<p>x</p>' }, slug: 'e2e-slug-' + k.slice(-4).toLowerCase() });
    const conflict = await consoleApi('POST', '/v2/console/publication/news', { title: { fa: 'نامک دوم e2e' }, bodyFormat: 'html', body: { fa: '<p>x</p>' }, slug: r1.slug });
    expect(conflict.status).toBe(409);
    expect(conflict.data.error).toBe('SLUG_EXISTS');
    const hidden = await pub.post('/news', { title: { fa: 'پنهان e2e ' + k.slice(-4) }, bodyFormat: 'html', body: { fa: '<p>x</p>' }, noindex: true });
    expect((await fetch(API_URL + '/v2/publication/public/sitemap.xml')).status).toBe(404);
    await open(page, '#/publication/news/settings');
    const form = root(page).locator('[data-settings-form="news"]');
    await form.locator('label.x-toggle', { hasText: 'صفحه‌های عمومی اخبار' }).click();
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'نشانی سایت')).toContainText('نشانی سایت لازم است');
    await field(form, 'نشانی سایت').locator('input').fill('https://site.e2e.example');
    await field(form, 'سبک نامک').locator('select').selectOption('latin');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    const sm = await fetch(API_URL + '/v2/publication/public/sitemap.xml');
    expect(sm.status).toBe(200);
    const xml = await sm.text();
    expect(xml).toContain('<loc>https://site.e2e.example/news/' + r1.slug + '</loc>');
    expect(xml).not.toContain(hidden.slug);
    const one = await fetch(API_URL + '/v2/publication/public/news/' + encodeURIComponent(r1.slug)).then((r) => r.json());
    expect(one.seo.metaTitle).toBe('نامک e2e ' + k.slice(-4));
    const latin = await pub.post('/news', { title: { fa: 'شب یلدا e2e ' + k.slice(-4) }, bodyFormat: 'html', body: { fa: '<p>x</p>' } });
    expect(latin.slug).toMatch(/^shb-ilda-e2e-/);
  });

  test.describe('in English', () => {
    test('English editor is LTR; 390px without sideways scroll on the news hub pages', async ({ page }) => {
      await open(page, '#/publication/news');
      await page.locator('.pl-seg-lang [data-locale="en"]').click();
      await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
      await root(page).getByRole('button', { name: 'New article' }).first().click();
      await settlePanel(page);
      await panel(page).locator('[data-translated="body"] [role="tab"][data-lang="en"]').click();
      await expect(panel(page).locator('[data-translated="body"] [role="tabpanel"][data-lang="en"] .eco-rt-body')).toHaveAttribute('dir', 'ltr');
      await expect(panel(page).locator('[data-rt="bold"]').first()).toHaveAttribute('title', 'Bold');
      await page.setViewportSize({ width: 390, height: 844 });
      for (const tab of ['', '/review', '/sources', '/log', '/reactions', '/settings']) {
        await open(page, '#/publication/news' + tab);
        await settle(page);
        await noHorizontalScroll(page);
      }
      await expect(root(page).locator('[data-review-bar]')).toHaveCount(0);
      await open(page, '#/publication/news/review');
      await expect(root(page).locator('[data-review-bar]')).toContainText('Approve');
      await page.setViewportSize({ width: 1280, height: 800 });
      await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
      await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
    });
  });
});
