import fs from 'node:fs';
import { expect, test } from '../../lib/fixtures';
import { consoleApi, deleteUser, devicePlayer, pub, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, pickRef, rows, savePanel, search, settlePanel } from '../../lib/console';

const k = uid();
const skey = 'e2e_' + k + '_sp';
const place = 'e2e_' + k + '_place';
const bkey = 'e2e_' + k + '_bn';
const made = { users: [] as string[] };
let P: Player;
let Q: Player;
let placeId = '';
let sponsorId = '';
let bannerId = '';

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await pub.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-sp-' + k, ('sp' + k).replace(/_/g, '').slice(0, 20));
  Q = await devicePlayer('e2e-sq-' + k, ('sq' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id, Q.id);
  await updateAccount(P, { lang_tag: 'fa' });
  await updateAccount(Q, { lang_tag: 'en' });
  const pl = await pub.post('/banner-lists', { key: place, name: { fa: 'جایگاه اسپانسر ' + k, en: 'Sponsor placement ' + k }, layout: 'single', maxItems: 2, rotateSeconds: 0 });
  placeId = pl.id;
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  const all = await pub.get('/banners?q=e2e_' + k + '&limit=100');
  for (const it of all.items || []) expect((await pub.del('/banners/' + it.id)).status).toBe(200);
  const lists = await pub.get('/banner-lists?q=e2e_' + k + '&limit=100');
  for (const it of lists.items || []) expect((await pub.del('/banner-lists/' + it.id)).status).toBe(200);
  const sps = await pub.get('/sponsors?q=e2e_' + k + '&limit=100');
  for (const it of sps.items || []) expect((await pub.del('/sponsors/' + it.id)).status).toBe(200);
});

test.describe('sponsors: console', () => {
  test('sponsor editor: key rule, links, private contact, contract window in the project zone, list row and locked key', async ({ page }) => {
    await open(page, '#/publication/sponsors');
    await expect(rows(page).filter({ hasText: 'آبمیوهٔ نوا' })).toHaveCount(1);
    await go(page, '#/publication/sponsors/new');
    await settlePanel(page);
    const p = panel(page);
    await expect(p).toContainText('اسپانسر جدید');
    await field(p, 'کلید').locator('input').fill('Bad Key');
    await fillTranslated(p, 'name', { fa: 'کافه ستاره ' + k, en: 'Star Cafe ' + k });
    await field(p, 'وب‌سایت').locator('input').fill('https://cafe.example.com/' + k);
    await field(p, 'رنگ برند').locator('input').fill('#0E7490');
    await p.locator('[data-rows-add="links"]').click();
    await p.locator('[data-rows="links"] [data-row="0"] [data-col="type"] select').selectOption('telegram');
    await p.locator('[data-rows="links"] [data-row="0"] [data-col="url"] input').fill('javascript:alert(1)');
    await field(p, 'ایمیل رابط').locator('input').fill('not-an-email');
    await savePanel(page).catch(() => null);
    await expect(field(p, 'کلید')).toContainText('کلید باید با حرف انگلیسی شروع شود');
    await expect(field(p, 'ایمیل رابط')).toContainText('نشانی ایمیل معتبر نیست');
    await expect(p.locator('[data-rows="links"] [data-row="0"]')).toContainText('نشانی باید با http');
    await field(p, 'کلید').locator('input').fill(skey);
    await p.locator('[data-rows="links"] [data-row="0"] [data-col="url"] input').fill('https://t.me/cafe' + k.replace(/[^a-z0-9]/g, ''));
    await field(p, 'ایمیل رابط').locator('input').fill('Ads@Cafe.Example.com');
    await field(p, 'تلفن رابط').locator('input').fill('۰۹۱۲۱۲۳۴۵۶۷');
    await expect(p).toContainText('این بخش فقط در کنسول دیده می‌شود');
    await field(p, /^شروع$/).locator('input').fill('2026-01-01T10:00');
    await expect(field(p, /^شروع$/).locator('[data-zone-shown]')).toContainText('Asia/Tehran');
    await savePanel(page);

    const sp = await byKey('sponsors', skey);
    expect(sp).toMatchObject({ color: '#0e7490', contactEmail: 'ads@cafe.example.com', contactPhone: '+989121234567', state: 'live', banners: 0, websiteUrl: 'https://cafe.example.com/' + k });
    expect(sp.links).toEqual([expect.objectContaining({ type: 'telegram' })]);
    expect(new Date(sp.startsAt).toISOString()).toBe('2026-01-01T06:30:00.000Z');
    sponsorId = sp.id;
    await search(page, skey);
    const row = rows(page).first();
    await expect(row).toContainText('کافه ستاره ' + k);
    await expect(row).toContainText('در حال نمایش');
    await expect(row.locator('[data-sp-banner-count]')).toHaveText('۰ بنر');
    await row.locator('td').first().click();
    await settlePanel(page);
    await expect(field(panel(page), 'کلید').locator('input')).toHaveJSProperty('readOnly', true);
    await expect(panel(page).locator('[data-sp-report]')).toContainText('هنوز هیچ بنر یا جایگاهی این اسپانسر را ندارد');

    const bad = await consoleApi('POST', '/v2/console/publication/sponsors', { key: skey + '_x', name: { fa: 'x' }, links: [{ type: 'fax', url: 'https://a.example.com' }] });
    expect(bad.status).toBe(400);
    expect(bad.data).toMatchObject({ error: 'INVALID_LINK_TYPE', field: 'links[0].type' });
  });

  test('sponsored banner: sponsor picker, sponsor action, list chip, placement sponsor and SPONSOR_REQUIRED', async ({ page }) => {
    await open(page, '#/publication/banners');
    await go(page, '#/publication/banners/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(bkey);
    await pickRef(p, 'جایگاه', 'جایگاه اسپانسر ' + k);
    await fillTranslated(p, 'title', { fa: 'قهوهٔ امروز ' + k, en: 'Coffee of the day ' + k });
    await field(p, /^کار$/).locator('select').selectOption('sponsor');
    await expect(p.locator('[data-bn-preview-action]')).toHaveText('زدن روی بنر وب‌سایت اسپانسر را باز می‌کند.');
    await savePanel(page).catch(() => null);
    await expect(field(p, /^اسپانسر$/)).toContainText('برای این دکمه اسپانسر را انتخاب کن');
    await pickRef(p, /^اسپانسر$/, 'کافه ستاره ' + k);
    await savePanel(page);
    const b = await byKey('banners', bkey);
    expect(b).toMatchObject({ sponsorId, sponsorKey: skey, action: { type: 'sponsor' } });
    bannerId = b.id;
    await search(page, bkey);
    await expect(rows(page).first().locator(`[data-bn-sponsor="${skey}"]`)).toHaveText('با حمایت کافه ستاره ' + k);

    await open(page, '#/publication/banner-lists');
    await go(page, '#/publication/banner-lists/' + placeId);
    await settlePanel(page);
    await pickRef(panel(page), 'اسپانسر جایگاه', 'آبمیوهٔ نوا');
    await savePanel(page);
    await search(page, place);
    await expect(rows(page).first().locator('[data-bl-sponsor]')).toHaveText('با حمایت آبمیوهٔ نوا');

    const none = await consoleApi('POST', '/v2/console/publication/banners', { key: bkey + '_x', listId: placeId, title: { fa: 'x' }, action: { type: 'sponsor' } });
    expect(none.status).toBe(400);
    expect(none.data.error).toBe('SPONSOR_REQUIRED');
  });

  test('game RPCs count each player once a day, the report, CSV download, contract end hides the banner, IN_USE and the audit log', async ({ page }) => {
    const list = await rpc(P, 'publication_banners', { placement: place, platform: 'web' });
    expect(list.data.items.map((x: any) => x.key)).toEqual([bkey]);
    expect(list.data.items[0].sponsor).toMatchObject({ key: skey, name: 'کافه ستاره ' + k });
    expect(list.data.items[0].action).toMatchObject({ type: 'sponsor', url: 'https://cafe.example.com/' + k });
    expect(list.data.placement.sponsor.key).toBe('star_sp_nova');
    const nova = (await byKey('sponsors', 'star_sp_nova')).id;
    for (let i = 0; i < 2; i++) {
      const v = await rpc(P, 'publication_banner_view', { bannerIds: [bannerId], platform: 'web' });
      expect(v.data.sponsorsCounted.sort()).toEqual(i === 0 ? [nova, sponsorId].sort() : []);
    }
    const c1 = await rpc(P, 'publication_banner_click', { bannerId, platform: 'web' });
    expect(c1.data).toMatchObject({ counted: true, sponsorCounted: true, action: { type: 'sponsor' } });
    expect((await rpc(P, 'publication_banner_click', { bannerId })).data.sponsorCounted).toBe(false);
    const sv = await rpc(Q, 'publication_sponsor_view', { sponsorIds: [sponsorId] });
    expect(sv.data.counted).toEqual([sponsorId]);
    const link = await rpc(Q, 'publication_sponsor_click', { sponsorId, link: 'telegram', lang: 'en' });
    expect(link.data).toMatchObject({ counted: true, url: expect.stringMatching(/^https:\/\/t\.me\//) });
    const mine = (await rpc(Q, 'publication_sponsors', { lang: 'en' })).data.items.find((x: any) => x.key === skey);
    expect(mine).toMatchObject({ name: 'Star Cafe ' + k, links: [expect.objectContaining({ type: 'telegram' })] });
    expect(mine.contactEmail).toBeUndefined();

    await open(page, '#/publication/sponsors');
    await go(page, '#/publication/sponsors/' + sponsorId);
    await settlePanel(page);
    const rep = panel(page).locator('[data-sp-report]');
    await expect(rep.locator('[data-sp-views]')).toContainText('۲');
    await expect(rep.locator('[data-sp-clicks]')).toContainText('۲');
    await expect(rep.locator('[data-sp-players]')).toContainText('۲');
    await expect(rep.locator('[data-sp-ctr]')).toContainText('۱۰۰٪');
    await expect(rep.locator(`[data-sp-banner="${bkey}"]`)).toContainText('۱ نمایش، ۱ کلیک');
    await expect(rep.locator('[data-sp-bars]')).toHaveAttribute('data-sp-bars', '14');
    await rep.locator('[data-sp-range] button', { hasText: '۷ روز' }).click();
    await expect(rep.locator('[data-sp-bars]')).toHaveAttribute('data-sp-bars', '7');
    const dl = page.waitForEvent('download');
    await rep.locator('[data-sp-csv]').click();
    const file = await dl;
    expect(file.suggestedFilename()).toMatch(new RegExp('^sponsor-' + skey + '-\\d{4}-\\d{2}-\\d{2}-\\d{4}-\\d{2}-\\d{2}\\.csv$'));
    const csv = fs.readFileSync(await file.path(), 'utf8');
    expect(csv.split('\n')[0]).toContain('day,views,clicks,players,ctr_percent');
    expect(csv).toContain('total,2,2,2,100.0');
    expect(csv).toContain(bkey + ',1,1,' + place + ',100.0');
    await page.keyboard.press('Escape');
    await search(page, skey);
    await expect(rows(page).first().locator('[data-sp-row-stats]')).toContainText('۲ نمایش، ۲ کلیک');

    const sp = await pub.get('/sponsors/' + sponsorId);
    await pub.put('/sponsors/' + sponsorId, { ...sp, startsAt: '2026-01-01T06:30:00Z', endsAt: new Date(Date.now() - 60_000).toISOString() });
    expect((await rpc(P, 'publication_banners', { placement: place })).data.items).toEqual([]);
    expect(JSON.stringify((await rpc(P, 'publication_banner_click', { bannerId })).data)).toContain('BANNER_NOT_AVAILABLE');
    expect(JSON.stringify((await rpc(Q, 'publication_sponsor_click', { sponsorId })).data)).toContain('SPONSOR_NOT_AVAILABLE');
    await go(page, '#/publication/banner-lists/' + placeId);
    await settlePanel(page);
    await panel(page).locator('[data-bn-try-player]').fill(P.username);
    await panel(page).locator('[data-bn-try-run]').click();
    await expect(panel(page).locator(`[data-bn-try-hidden="${bkey}"]`)).toContainText('قرارداد اسپانسرش فعال نیست');
    await page.keyboard.press('Escape');

    const used = await consoleApi('DELETE', '/v2/console/publication/sponsors/' + sponsorId);
    expect(used.status).toBe(409);
    expect(used.data.error).toBe('IN_USE');
    const audit = sql(`SELECT string_agg(DISTINCT action, ',' ORDER BY action) FROM console_audit_log WHERE message LIKE 'publication % sponsor%' AND create_time > now() - interval '10 minutes'`);
    expect(audit.split(',')).toEqual(expect.arrayContaining(['CREATE', 'UPDATE']));
  });

  test('sidebar, page search, English and 390px cards in dark mode without sideways scroll', async ({ page, spawn }) => {
    await open(page, '#/publication/sponsors');
    await expect(page.locator('.pl-side a.pl-nav[href^="#/publication/sponsors"] svg')).toHaveCount(1);
    await expect(page.locator('.pl-side a.pl-nav[href^="#/publication/sponsors"]')).toContainText('اسپانسرها');
    await go(page, '#/publication/faqs');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('اسپانسرها');
    await expect(page.locator('.pl-pal-row').first()).toContainText('اسپانسرها');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/publication\/sponsors/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/publication/sponsors?q=' + skey);
    const en = rows(page).filter({ hasText: 'Star Cafe ' + k });
    await expect(en).toHaveCount(1);
    await expect(en).toContainText('Ended');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);

    const m = await spawn({ signedIn: true, mode: 'dark', mobile: true });
    await open(m, '#/publication/sponsors');
    await expect(m.locator('html')).toHaveClass(/dark/);
    await noHorizontalScroll(m);
    await go(m, '#/publication/sponsors/' + sponsorId);
    await settlePanel(m);
    await expect(panel(m).locator('[data-sp-report]')).toBeVisible();
    await noHorizontalScroll(m);
  });
});
