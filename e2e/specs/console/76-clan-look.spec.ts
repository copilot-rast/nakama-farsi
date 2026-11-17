import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, rpc, uid, type Player } from '../../lib/api';
import { cln, createClan, rawSettings, removeClan, restoreSettings } from '../../lib/clan';
import { noHorizontalScroll } from '../../lib/console';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const short = k.slice(-7);
const part = 'e2e_' + k;
const EM = part + '_em';
const CV = part + '_cv';
const TAG = part + '_friendly';
const PNG = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
const F: Record<string, any> = {};
const users: string[] = [];
let raw = '';

test.describe.configure({ mode: 'serial' });

const root = (page: Page) => page.locator('#clan-root');
const panel = (page: Page) => page.locator('[data-clan=panel] [role=dialog]');
const toasts = (page: Page) => page.locator('[data-clan=toasts] .eco-notice');

async function openPage(page: Page, sub: string) {
  await page.goto(CONSOLE_URL + '/#/' + sub);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await expect(root(page).locator(`[data-clan-page="${sub.split('?')[0]}"]`)).toBeVisible();
  await expect(root(page).getByText('در حال بارگذاری…')).toHaveCount(0);
}

async function openClan(page: Page) {
  await openPage(page, 'clans/list');
  await root(page).locator('input[name="q"]').fill('e2e' + short);
  await root(page).locator(`tr[data-clan-row="${F.clan}"]`).click();
  await expect(panel(page).locator('[data-clan-form="look"]')).toBeVisible();
}

test.beforeAll(async () => {
  raw = rawSettings('clan_settings');
  await cln.setSettings((s) => {
    s.enabled = true; s.create.enabled = true; s.create.cost = {}; s.create.minTrophies = 0; s.events.publish = false; s.events.notify = false;
    s.profile.changeCooldownMinutes = 0; s.profile.requireEmblem = false; s.profile.colors = []; s.profile.maxTags = 3; s.profile.mottoMax = 60;
    s.profile.tags = [...(s.profile.tags || []), { key: TAG, name: { fa: 'دوستانه ' + short, en: 'Friendly ' + short } }];
    s.uploads = { enabled: true, needsApproval: true, maxBytes: 300 * 1024, types: ['png', 'jpg', 'webp'] };
  });
  const p: Player = await devicePlayer('e2e-clook-a-' + k);
  users.push(p.id);
  F.a = p;
  F.clan = (await createClan(p, 'e2e' + short + ' Owls')).clanId;
  F.country = sql(`SELECT code FROM base_country WHERE is_active ORDER BY sort, code LIMIT 1`).trim();
});

test.afterAll(async () => {
  await removeClan(F.clan);
  sql(`DELETE FROM clan_art WHERE key LIKE '%${part}%';`);
  for (const id of users) await deleteUser(id);
  if (raw) {
    await cln.raw('PUT', '/settings', JSON.parse(raw));
    restoreSettings('clan_settings', raw);
  }
});

test.describe('clan look console', () => {
  test('gallery: an emblem with an uploaded file, a cover by URL, a missing image is refused, then edited', async ({ page }) => {
    await openPage(page, 'clans/gallery');
    await expect(page.locator('[data-clan-group="clans"] a[href="#/clans/uploads"]')).toBeAttached();
    await root(page).locator('[data-clan-action="new-art"]').click();
    let p = panel(page);
    await p.locator('select[name="kind"]').selectOption('emblem');
    await p.locator('input[name="key"]').fill(EM);
    await p.locator('input[name="name-fa"]').fill('عقاب ' + short);
    await p.locator('[data-clan-action="save-art"]').click();
    await expect(p.locator('.eco-notice')).toContainText('تصویر لازم است');
    await p.locator('input[name="imageUrl-file"]').setInputFiles({ name: 'emblem.png', mimeType: 'image/png', buffer: Buffer.from(PNG, 'base64') });
    await expect(p.locator('input[name="imageUrl"]')).toHaveValue(/^https?:\/\/.+/);
    await expect(p.locator('[data-clan-preview="imageUrl"] img')).toBeAttached();
    await p.locator('input[name="minLevel"]').fill('۲');
    await p.locator('[data-clan-action="save-art"]').click();
    await expect(p).toHaveCount(0);
    const card = root(page).locator(`[data-clan-art="emblem:${EM}"]`);
    await expect(card).toContainText('عقاب ' + short);
    await expect(card).toContainText('از سطح ۲');

    await root(page).locator('[data-clan-action="new-art"]').click();
    p = panel(page);
    await p.locator('select[name="kind"]').selectOption('cover');
    await p.locator('input[name="key"]').fill(CV);
    await p.locator('input[name="name-fa"]').fill('کوهستان ' + short);
    await p.locator('input[name="imageUrl"]').fill('https://example.com/' + CV + '.png');
    await p.locator('[data-clan-action="save-art"]').click();
    await expect(p).toHaveCount(0);
    await root(page).locator('[data-clan-kind="cover"]').click();
    await expect(root(page).locator(`[data-clan-art="emblem:${EM}"]`)).toHaveCount(0);
    await root(page).locator(`[data-clan-art="cover:${CV}"]`).click();
    p = panel(page);
    await p.locator('input[name="name-en"]').fill('Mountain ' + short);
    await p.locator('[data-clan-action="save-art"]').click();
    await expect(p).toHaveCount(0);
    const list = (await cln.get('/art')).items;
    expect(list.find((a: any) => a.key === EM)).toMatchObject({ kind: 'emblem', minLevel: 2, isActive: true, isDefault: false });
    expect(list.find((a: any) => a.key === CV)).toMatchObject({ kind: 'cover', imageUrl: 'https://example.com/' + CV + '.png', name: { fa: 'کوهستان ' + short, en: 'Mountain ' + short } });
    F.em = list.find((a: any) => a.key === EM);
  });

  test('admin edits a clan look past the level lock; a used emblem cannot be deleted; reset look after a confirm', async ({ page }) => {
    await openClan(page);
    const f = panel(page).locator('[data-clan-form="look"]');
    await f.locator('select[name="emblemKey"]').selectOption(EM);
    await f.locator('select[name="coverKey"]').selectOption(CV);
    await f.locator('input[name="look-color"]').fill('#12AB34');
    await f.locator('input[name="motto"]').fill('با هم ' + short);
    if (F.country) await f.locator('[name="country"]').selectOption(F.country);
    await f.locator('select[name="lang"]').selectOption('fa');
    await f.locator('[data-clan-tags] label', { hasText: 'دوستانه ' + short }).locator('input').check();
    await panel(page).locator('[data-clan-action="save-look"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    await expect(panel(page).locator('[data-clan-value="motto"]')).toHaveText('«با هم ' + short + '»');
    await expect(panel(page).locator('[data-clan-look]')).toContainText('دوستانه ' + short);
    await expect(panel(page).locator('[data-clan-emblem] img')).toHaveAttribute('src', F.em.imageUrl);
    const d = await cln.get('/list/' + F.clan);
    expect(d.clan).toMatchObject({ emblemKey: EM, coverKey: CV, color: '#12AB34', motto: 'با هم ' + short, lang: 'fa', tags: [TAG] });
    if (F.country) expect(d.clan.country).toBe(F.country);

    await openPage(page, 'clans/gallery');
    await root(page).locator(`[data-clan-art="emblem:${EM}"]`).click();
    await panel(page).locator('[data-clan-action="delete-art"]').click();
    await page.locator('[data-clan-confirm="ok"]').click();
    await expect(panel(page).locator('.eco-notice')).toContainText('۱ کلن از این تصویر استفاده می‌کنند');
    await page.keyboard.press('Escape');

    await openClan(page);
    await panel(page).locator('[data-clan-action="reset-look"]').click();
    await page.locator('[data-clan-confirm="ok"]').click();
    await expect(toasts(page).last()).toContainText('ظاهر کلن بازگردانی شد');
    await expect(panel(page).locator('[data-clan-value="motto"]')).toHaveCount(0);
    const r = await cln.get('/list/' + F.clan);
    expect(r.clan.motto).toBe('');
    expect(r.clan.color).toBe('');
    expect(r.clan.emblemKey).not.toBe(EM);
    expect(r.clan.coverKey).not.toBe(CV);
  });

  test('settings: profile tags, words and colours and the upload rules; a bad size names its field; the original comes back', async ({ page }) => {
    const orig = await cln.settings();
    await openPage(page, 'clans/settings');
    const box = root(page);
    const before = await box.locator('[data-clan-tag-row]').count();
    await box.locator('[data-clan-action="add-tag"]').click();
    const row = box.locator('[data-clan-tag-row]').last();
    await row.locator('input[name="tag-key"]').fill(part + '_pro');
    await row.locator('input[name="tag-name-fa"]').fill('حرفه‌ای');
    await row.locator('input[name="tag-name-en"]').fill('Pro');
    await box.locator('textarea[name="profile.blockedWords"]').fill('badword\nبدواژه');
    await box.locator('input[name="profile.colors"]').fill('#e53935، #1E88E5');
    await box.locator('input[name="profile.mottoMax"]').fill('۴۰');
    await box.locator('input[name="profile.maxTags"]').fill('۲');
    await box.locator('[data-clan-types] label', { hasText: 'WEBP' }).locator('input').uncheck();
    await box.locator('input[name="uploads.maxKb"]').fill('0');
    await box.locator('[data-clan-save="clans/settings"]').click();
    await expect(box.locator('[data-clan-msg]')).toContainText('uploads.maxBytes');
    await box.locator('input[name="uploads.maxKb"]').fill('۲۰۰');
    await box.locator('[data-clan-save="clans/settings"]').click();
    await expect(toasts(page).last()).toContainText('ذخیره شد');
    const s = await cln.settings();
    expect(s.profile.tags.length).toBe(before + 1);
    expect(s.profile.tags[before]).toEqual({ key: part + '_pro', name: { fa: 'حرفه‌ای', en: 'Pro' } });
    expect(s.profile).toMatchObject({ blockedWords: ['badword', 'بدواژه'], colors: ['#E53935', '#1E88E5'], mottoMax: 40, maxTags: 2 });
    expect(s.uploads).toMatchObject({ enabled: true, maxBytes: 200 * 1024, types: ['png', 'jpg'] });
    await cln.put('/settings', orig);
    await page.reload();
    await expect(root(page).locator('[data-clan-tag-row]')).toHaveCount(before);
  });

  test('uploads queue: a pending emblem shows with its preview and badge, is approved; the next one is rejected with a reason', async ({ page }) => {
    const up = await rpc(F.a, 'clan_image_upload', { kind: 'emblem', data: PNG });
    expect(up.status, JSON.stringify(up.data)).toBe(200);
    expect(up.data.status).toBe('pending');
    await openPage(page, 'clans/uploads');
    await expect(page.locator('[data-clan-nav="clans/uploads"] .eco-nav-badge')).toBeAttached();
    await expect(page.locator('.pl-side a.pl-nav[href="#/clans/uploads"] .pl-badge')).toBeVisible();
    const card = root(page).locator(`[data-clan-upload="${up.data.uploadId}"]`);
    await expect(card).toContainText('e2e' + short + ' Owls');
    await expect(card.locator('img')).toBeVisible();
    await card.locator('[data-clan-action="approve"]').click();
    await expect(toasts(page).last()).toContainText('تصویر تأیید شد');
    await expect(card).toHaveCount(0);
    expect((await cln.get('/uploads?status=approved')).items.find((u: any) => u.id === up.data.uploadId)).toBeTruthy();
    const after = (await cln.get('/list/' + F.clan)).clan;
    expect(after.emblemUrl).toBeTruthy();
    expect(after.pendingUploads.emblem).toBe(false);

    const two = await rpc(F.a, 'clan_image_upload', { kind: 'cover', data: PNG });
    expect(two.status, JSON.stringify(two.data)).toBe(200);
    await page.reload();
    const c2 = root(page).locator(`[data-clan-upload="${two.data.uploadId}"]`);
    await c2.locator('[data-clan-action="reject"]').click();
    await page.locator('[data-clan=confirm] textarea[name="reason"]').fill('تصویر مناسب نیست');
    await page.locator('[data-clan-confirm="ok"]').click();
    await expect(toasts(page).last()).toContainText('تصویر رد شد');
    await root(page).locator('[data-clan-status="rejected"]').click();
    await expect(root(page).locator(`[data-clan-upload="${two.data.uploadId}"] [data-clan-reason]`)).toHaveText('تصویر مناسب نیست');
    const r = (await cln.get('/uploads?status=rejected')).items.find((u: any) => u.id === two.data.uploadId);
    expect(r).toMatchObject({ status: 'rejected', reason: 'تصویر مناسب نیست', kind: 'cover' });
    const again = await cln.raw('POST', '/uploads/' + two.data.uploadId + '/approve', {});
    expect(again.status).toBe(409);
    expect(again.data.error).toBe('CLAN_UPLOAD_NOT_PENDING');
    const audit = sql(`SELECT string_agg(DISTINCT regexp_replace(message, '[0-9a-f-]{36}', 'id', 'g'), ',' ORDER BY regexp_replace(message, '[0-9a-f-]{36}', 'id', 'g')) FROM console_audit_log WHERE (message LIKE '%${F.clan}/reset-look' OR message LIKE '%clans/uploads/${up.data.uploadId}/approve' OR message LIKE '%clans/uploads/${two.data.uploadId}/reject') AND create_time > now() - interval '10 minutes'`);
    expect(audit).toBe('base invoke clans/list/id/reset-look,base invoke clans/uploads/id/approve,base invoke clans/uploads/id/reject');
  });

  test('a gallery image is deleted; Ctrl K and 390px English dark without sideways scroll', async ({ page, spawn }) => {
    await openPage(page, 'clans/gallery');
    await root(page).locator(`[data-clan-art="cover:${CV}"]`).click();
    await panel(page).locator('[data-clan-action="delete-art"]').click();
    await page.locator('[data-clan-confirm="ok"]').click();
    await expect(root(page).locator(`[data-clan-art="cover:${CV}"]`)).toHaveCount(0);
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('تصویرهای ارسالی');
    await expect(page.locator('.pl-pal-row').first()).toContainText('تصویرهای ارسالی');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/clans\/uploads/);

    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    for (const [sub, title] of [['gallery', 'Emblems and covers'], ['uploads', 'Uploaded images'], ['settings', 'Clan settings']]) {
      await m.goto(CONSOLE_URL + '/#/clans/' + sub);
      await expect(m.locator(`#clan-root [data-clan-page="clans/${sub}"] h1`)).toHaveText(title);
      await expect(m.getByText('Loading…')).toHaveCount(0);
      await noHorizontalScroll(m);
    }
    await m.goto(CONSOLE_URL + '/#/clans/uploads');
    await m.locator('#clan-root [data-clan-status="all"]').click();
    await expect(m.locator('#clan-root [data-clan-upload]').first()).toBeVisible();
    await noHorizontalScroll(m);
    await m.goto(CONSOLE_URL + '/#/clans/list');
    await m.locator('#clan-root input[name="q"]').fill('e2e' + short);
    await m.locator(`#clan-root tr[data-clan-row="${F.clan}"]`).click();
    await expect(m.locator('[data-clan=panel] [data-clan-form="look"]')).toContainText('Reset look');
    await noHorizontalScroll(m);
  });
});
