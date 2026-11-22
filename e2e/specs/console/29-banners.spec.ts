import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { FIXTURES } from '../../lib/env';
import { base, consoleApi, deleteUser, devicePlayer, pub, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const place = 'e2e_' + k + '_place';
const bkey = 'e2e_' + k + '_bn';
const made = { users: [] as string[], tags: [] as string[] };
let P: Player;
let placeId = '';
let tagId = '';

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await pub.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
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
  P = await devicePlayer('e2e-bn-' + k, ('bn' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
  const tag = await base.post('/tags', { keyType: 'user', key: 'e2e_' + k + '_fans', name: { fa: 'هواداران ' + k, en: 'Fans ' + k } });
  tagId = tag.id;
  made.tags.push(tag.id);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  const all = await pub.get('/banners?q=e2e_' + k + '&limit=100');
  for (const it of all.items || []) expect((await pub.del('/banners/' + it.id)).status).toBe(200);
  const lists = await pub.get('/banner-lists?q=e2e_' + k + '&limit=100');
  for (const it of lists.items || []) expect((await pub.del('/banner-lists/' + it.id)).status).toBe(200);
  for (const id of made.tags) await base.del('/tags/' + id);
});

test.describe('banners: console', () => {
  test('placement editor: key rule, layout, list row and locked key', async ({ page }) => {
    await open(page, '#/publication/banner-lists');
    await expect(rows(page).filter({ hasText: 'چرخ‌فلک منوی اصلی' })).toHaveCount(1);
    await go(page, '#/publication/banner-lists/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('جایگاه جدید');
    await field(p, 'کلید').locator('input').fill('Bad Key');
    await fillTranslated(p, 'name', { fa: 'جایگاه آزمایشی ' + k, en: 'Test placement ' + k });
    await savePanel(page).catch(() => null);
    await expect(field(p, 'کلید')).toContainText('کلید باید با حرف انگلیسی شروع شود');
    await field(p, 'کلید').locator('input').fill(place);
    await field(p, 'بیشترین تعداد').locator('input').fill('3');
    await field(p, 'چرخش (ثانیه)').locator('input').fill('4');
    await savePanel(page);
    const pl = await byKey('banner-lists', place);
    expect(pl).toMatchObject({ layout: 'carousel', maxItems: 3, rotateSeconds: 4, isActive: true, banners: 0 });
    placeId = pl.id;
    await search(page, place);
    const row = rows(page).first();
    await expect(row).toContainText('جایگاه آزمایشی ' + k);
    await expect(row).toContainText('چرخ‌فلک');
    await expect(row).toContainText('هر ۴ ثانیه، تا ۳ بنر');
    await expect(row.locator('[data-bn-count]')).toHaveText('۰ فعال از ۰');
    await row.locator('td').first().click();
    await settlePanel(page);
    await expect(field(panel(page), 'کلید').locator('input')).toHaveJSProperty('readOnly', true);
    await expect(panel(page).locator('[data-bl-order]')).toContainText('هنوز بنری در این جایگاه نیست');
  });

  test('banner editor: live preview, per-language image upload, product action, schedule in the project zone', async ({ page }) => {
    await open(page, '#/publication/banners');
    await go(page, '#/publication/banners/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('بنر جدید');
    await field(p, 'کلید').locator('input').fill(bkey);
    await pickIn(field(p, 'جایگاه'), 'جایگاه آزمایشی ' + k);
    await field(p, 'اولویت').locator('input').fill('20');
    const preview = p.locator('[data-bn-preview]');
    await fillTranslated(p, 'title', { fa: 'حراج کهکشانی ' + k, en: 'Galaxy sale ' + k });
    await expect(preview.locator('[data-bn-preview-title]')).toHaveText('حراج کهکشانی ' + k);
    await preview.locator('[data-bn-langs] button', { hasText: 'English' }).click();
    await expect(preview.locator('[data-bn-preview-title]')).toHaveText('Galaxy sale ' + k);
    await preview.locator('[data-bn-langs] button', { hasText: 'فارسی' }).click();

    const image = field(p, /^تصویر$/);
    const upload = page.waitForResponse((r) => r.url().includes('/v2/console/economy/uploads') && r.request().method() === 'POST');
    await image.locator('[role="tabpanel"][data-lang="fa"] input[type="file"]').setInputFiles(path.join(FIXTURES, 'logo.svg'));
    expect((await upload).ok()).toBe(true);
    await expect(image.locator('[role="tabpanel"][data-lang="fa"] input[type="text"]')).toHaveValue(/\/v2\/economy\/files\/.+\.svg$/);
    await expect(preview.locator('img.eco-bn-img')).toBeVisible();
    await image.locator('[role="tab"][data-lang="en"]').click();
    await image.locator('[role="tabpanel"][data-lang="en"] input[type="text"]').fill('https://cdn.example.com/en-' + k + '.png');

    await field(p, /^کار$/).locator('select').selectOption('product');
    await pickIn(field(p, 'محصول'), 'سبد کهکشانی');
    await fillTranslated(p, 'action.label', { fa: 'همین حالا بخر' });
    await expect(preview.locator('[data-bn-preview-cta]')).toHaveText('همین حالا بخر');
    await expect(preview.locator('[data-bn-preview-action]')).toContainText('star_skin_galaxy');

    await expect(p.locator('[data-bn-zone]')).toContainText('Asia/Tehran');
    await field(p, /^شروع$/).locator('input').fill('2030-01-01T10:00');
    await expect(field(p, /^شروع$/).locator('[data-zone-shown]')).toContainText('Asia/Tehran');
    await field(p, 'از ساعت').locator('input').fill('09:00');
    await savePanel(page).catch(() => null);
    await expect(field(p, 'تا ساعت')).toContainText('هر دو ساعت را بده');
    await field(p, 'تا ساعت').locator('input').fill('21:30');
    await field(p, 'روزهای هفته').locator('label', { hasText: 'جمعه' }).click();
    await savePanel(page);

    const b = await byKey('banners', bkey);
    expect(b).toMatchObject({ listId: placeId, priority: 20, state: 'scheduled', dayStart: '09:00', dayEnd: '21:30', days: [5],
      action: { type: 'product', value: 'star_skin_galaxy', label: { fa: 'همین حالا بخر' } } });
    expect(new Date(b.startsAt).toISOString()).toBe('2030-01-01T06:30:00.000Z');
    expect(b.imageUrl.en).toBe('https://cdn.example.com/en-' + k + '.png');
    expect(b.imageUrl.fa).toMatch(/\.svg$/);
    await search(page, bkey);
    const row = rows(page).first();
    await expect(row).toContainText('حراج کهکشانی ' + k);
    await expect(row).toContainText('زمان‌بندی‌شده');
    await expect(row).toContainText('اولویت ۲۰');
    await expect(row.locator('[data-bn-target]')).toContainText('جمعه');
    await expect(row.locator('[data-bn-target]')).toContainText('ساعت ۰۹:۰۰ تا ۲۱:۳۰');
    await row.locator('td').first().click();
    await settlePanel(page);
    await expect(field(panel(page), /^شروع$/).locator('input')).toHaveValue('۱۴۰۸/۱۰/۱۲ ۱۰:۰۰');

    const bad = await consoleApi('POST', '/v2/console/publication/banners', { key: bkey + '_x', listId: placeId, title: { fa: 'x' }, action: { type: 'url', url: { fa: 'javascript:alert(1)' } } });
    expect(bad.status).toBe(400);
    expect(bad.data.error).toBe('INVALID_URL');
  });

  test('targeting, the player check, counters from the game RPCs and the audit log', async ({ page }) => {
    const live = await pub.post('/banners', { key: bkey + '_open', listId: placeId, priority: 50, title: { fa: 'برای همه ' + k }, action: { type: 'screen', value: 'tasks' } });
    await pub.post('/banners', { key: bkey + '_fans', listId: placeId, priority: 40, title: { fa: 'فقط هواداران ' + k }, audienceTagIds: [tagId] });

    await open(page, '#/publication/banner-lists');
    await go(page, '#/publication/banner-lists/' + placeId);
    await settlePanel(page);
    const p = panel(page);
    await expect(p.locator('[data-bl-item]')).toHaveCount(3);
    await expect(p.locator('[data-bl-item]').first()).toHaveAttribute('data-bl-item', bkey + '_open');
    await p.locator('[data-bn-try-player]').fill(P.username);
    await p.locator('[data-bn-try-run]').click();
    const out = p.locator('[data-bn-try-out]');
    await expect(out.locator('[data-bn-try-who]')).toContainText('۱ بنر');
    await expect(out.locator(`[data-bn-try-shown="${bkey}_open"]`)).toBeVisible();
    await expect(out.locator(`[data-bn-try-hidden="${bkey}_fans"]`)).toContainText('برچسب لازم را ندارد');
    await expect(out.locator(`[data-bn-try-hidden="${bkey}"]`)).toContainText('هنوز شروع نشده');
    await page.keyboard.press('Escape');

    const list = await rpc(P, 'publication_banners', { placement: place, platform: 'web' });
    expect(list.data.items.map((x: any) => x.key)).toEqual([bkey + '_open']);
    expect(list.data.placement).toMatchObject({ layout: 'carousel', rotateSeconds: 4 });
    const fans = await byKey('banners', bkey + '_fans');
    const v1 = await rpc(P, 'publication_banner_view', { bannerIds: [live.id, fans.id], platform: 'web' });
    expect(v1.data).toMatchObject({ counted: [live.id], skipped: [fans.id] });
    const v2 = await rpc(P, 'publication_banner_view', { bannerIds: [live.id], platform: 'web' });
    expect(v2.data.counted).toEqual([]);
    const c1 = await rpc(P, 'publication_banner_click', { bannerId: live.id, platform: 'web' });
    expect(c1.data).toMatchObject({ counted: true, action: { type: 'screen', value: 'tasks' } });
    expect((await rpc(P, 'publication_banner_click', { bannerId: live.id })).data.counted).toBe(false);
    const sneaky = await rpc(P, 'publication_banner_click', { bannerId: fans.id });
    expect(JSON.stringify(sneaky.data)).toContain('BANNER_NOT_AVAILABLE');

    await base.put('/players/' + P.id + '/tags', { tagIds: [tagId] });
    const after = await rpc(P, 'publication_banners', { placement: place });
    expect(after.data.items.map((x: any) => x.key)).toEqual([bkey + '_open', bkey + '_fans']);

    await open(page, '#/publication/banners');
    await go(page, '#/publication/banners/' + live.id);
    await settlePanel(page);
    const stats = panel(page).locator('[data-bn-stats]');
    await expect(stats.locator('[data-bn-stat-views]')).toHaveText('۱ نمایش در ۱۴ روز');
    await expect(stats.locator('[data-bn-stat-clicks]')).toHaveText('۱ کلیک');
    await expect(stats).toContainText('نرخ کلیک ۱۰۰٪');
    await page.keyboard.press('Escape');
    await search(page, bkey + '_open');
    await expect(rows(page).first().locator('[data-bn-row-stats]')).toContainText('۱ نمایش، ۱ کلیک');

    await open(page, '#/publication/banner-lists');
    const used = await consoleApi('DELETE', '/v2/console/publication/banner-lists/' + placeId);
    expect(used.status).toBe(409);
    expect(used.data.error).toBe('IN_USE');
    await pub.put('/banners/' + live.id, { key: bkey + '_open', listId: placeId, priority: 55, title: { fa: 'برای همه ' + k }, action: { type: 'screen', value: 'tasks' } });
    const audit = sql(`SELECT string_agg(DISTINCT action, ',' ORDER BY action) FROM console_audit_log WHERE message LIKE 'publication % banner%' AND create_time > now() - interval '5 minutes'`);
    expect(audit.split(',')).toEqual(expect.arrayContaining(['CREATE', 'UPDATE']));
  });

  test('sidebar, page search, English and 390px cards in dark mode without sideways scroll', async ({ page, spawn }) => {
    await open(page, '#/publication/banners');
    await expect(page.locator('.pl-side a.pl-nav[href^="#/publication/banners"] svg')).toHaveCount(1);
    await expect(root(page).locator('[data-economy-tabs="banners"] a[data-economy-tab="banner-lists"]')).toContainText('جایگاه‌های بنر');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('جایگاه‌های بنر');
    await expect(page.locator('.pl-pal-row').first()).toContainText('جایگاه‌های بنر');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/publication\/banners\/placements/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/publication/banners?q=' + bkey);
    const en = rows(page).filter({ hasText: 'Galaxy sale ' + k });
    await expect(en).toHaveCount(1);
    await expect(en).toContainText('Scheduled');
    await expect(en).toContainText('Priority 20');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);

    const m = await spawn({ signedIn: true, mode: 'dark', mobile: true });
    await open(m, '#/publication/banners');
    await expect(m.locator('html')).toHaveClass(/dark/);
    await noHorizontalScroll(m);
    await go(m, '#/publication/banner-lists');
    await noHorizontalScroll(m);
    await go(m, '#/publication/banners/' + (await byKey('banners', bkey)).id);
    await settlePanel(m);
    await expect(panel(m).locator('[data-bn-preview]')).toBeVisible();
    await noHorizontalScroll(m);
  });
});
