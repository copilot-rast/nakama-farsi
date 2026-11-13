import crypto from 'node:crypto';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, deleteUser, devicePlayer, eco, pub, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { restoreConfig, sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, openSettings, panel, pickRef, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const catKey = 'e2e_' + k + '_cat';
const ruleKey = 'e2e_' + k + '_rule';
const evKey = 'e2e_' + k + '_ev';
const made = { users: [] as string[], tags: [] as string[], types: [] as string[] };
let A: Player;
let B: Player;
let C: Player;
let tagId = '';
let catId = '';

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string, api = pub) {
  const res = await api.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

function serviceAccount(project: string) {
  const { privateKey } = crypto.generateKeyPairSync('rsa', { modulusLength: 2048 });
  return JSON.stringify({ type: 'service_account', project_id: project, private_key_id: 'e2e', private_key: privateKey.export({ type: 'pkcs8', format: 'pem' }), client_email: 'push@' + project + '.iam.gserviceaccount.com' }, null, 2);
}

test.beforeAll(async () => {
  A = await devicePlayer('e2e-nt-a-' + k, ('nta' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-nt-b-' + k, ('ntb' + k).replace(/_/g, '').slice(0, 20));
  C = await devicePlayer('e2e-nt-c-' + k, ('ntc' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id, C.id);
  for (const p of [A, B, C]) await updateAccount(p, { lang_tag: 'fa' });
  const tag = await base.post('/tags', { keyType: 'user', key: 'e2e_' + k + '_push', name: { fa: 'پوش ' + k, en: 'Push ' + k } });
  tagId = tag.id;
  made.tags.push(tag.id);
  for (const p of [A, B, C]) await base.put('/players/' + p.id + '/tags', { tagIds: [tagId] });
  expect((await rpc(A, 'publication_push_register', { token: 'sandbox-web-e2e-a-' + k, platform: 'web', appVersion: '1.4.0' })).status).toBe(200);
  expect((await rpc(B, 'publication_push_register', { token: 'sandbox-dead-e2e-b-' + k, platform: 'android', appVersion: '1.4.0' })).status).toBe(200);
  expect((await rpc(C, 'publication_push_register', { token: 'sandbox-web-e2e-c-' + k, platform: 'web', appVersion: '1.4.0' })).status).toBe(200);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  const rules = await pub.get('/auto-rules?q=e2e_' + k + '&limit=100');
  for (const it of rules.items || []) expect((await pub.del('/auto-rules/' + it.id)).status).toBe(200);
  for (const id of made.types) await eco.del('/event-types/' + id);
  sql(`DELETE FROM publication_notification WHERE category_id IN (SELECT id FROM publication_notify_category WHERE key LIKE 'e2e\\_${k}%')`);
  const cats = await pub.get('/notify-categories?q=e2e_' + k + '&limit=100');
  for (const it of cats.items || []) expect((await pub.del('/notify-categories/' + it.id)).status).toBe(200);
  for (const id of made.tags) await base.del('/tags/' + id);
  restoreConfig();
});

test.describe('notifications: console', () => {
  test('categories: seeded list, a new category off by default, the always-on rule; push settings with the log service and a masked FCM key', async ({ page }) => {
    await open(page, '#/publication/notify-categories');
    const system = rows(page).filter({ hasText: 'پیام‌های مهم' });
    await expect(system).toHaveCount(1);
    await expect(system).toContainText('همیشه روشن');
    await go(page, '#/publication/notify-categories/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(catKey);
    await fillTranslated(p, 'name', { fa: 'تخفیف آزمایشی ' + k, en: 'Test deals ' + k });
    await p.locator('.x-toggle', { hasText: 'پیش‌فرض روشن' }).click();
    await savePanel(page);
    const locked = await consoleApi('POST', '/v2/console/publication/notify-categories', { key: catKey + '_x', name: { fa: 'x' }, canOptOut: false, defaultOn: false });
    expect(locked.status).toBe(400);
    expect(locked.data).toMatchObject({ error: 'CATEGORY_ALWAYS_ON', field: 'defaultOn' });
    const cat = await byKey('notify-categories', catKey);
    expect(cat).toMatchObject({ canOptOut: true, defaultOn: false });
    await pub.put('/notify-categories/' + cat.id, { ...cat, defaultOn: true });
    catId = cat.id;
    expect((await rpc(C, 'publication_notify_prefs_set', { prefs: { [catKey]: false } })).status).toBe(200);
    await search(page, catKey);
    await expect(rows(page).first().locator('[data-nc-counts]')).toContainText('۱ نفر خاموش کرده‌اند');

    await openSettings(page, 'push');
    const logCard = root(page).locator('[data-provider="log"]');
    await expect(logCard).toContainText('ثبت در گزارش (آزمایشی)');
    await expect(logCard).toContainText('این سرویس هیچ‌وقت پیام واقعی نمی‌فرستد');
    const bad = await consoleApi('POST', '/v2/console/publication/providers', { channel: 'push', kind: 'fcm', config: { serviceAccount: '{"type":"authorized_user"}' } });
    expect(bad.status).toBe(400);
    expect(bad.data).toMatchObject({ error: 'INVALID_SERVICE_ACCOUNT', field: 'serviceAccount' });
    const fcm = await pub.post('/providers', { channel: 'push', kind: 'fcm', isActive: false, sandbox: true, config: { serviceAccount: serviceAccount('e2e-proj-' + k.replace(/_/g, '-')) } });
    expect(JSON.stringify(fcm)).not.toContain('PRIVATE KEY');
    expect(fcm.config.serviceAccount).toBe('********');
    await page.reload();
    await expect(root(page).locator('[data-provider="fcm"] [data-pv-project]')).toHaveText('پروژهٔ e2e-proj-' + k.replace(/_/g, '-'));
    await expect(root(page).locator('[data-provider="fcm"] textarea[data-secret-json]')).toHaveValue('********');
    const testRes = await pub.post('/providers/' + (await pub.get('/providers')).items.find((x: any) => x.kind === 'log').id + '/test', { to: A.username, message: 'سلام', subject: 'آزمایش' });
    expect(testRes).toMatchObject({ ok: true, sandbox: true, sent: 1 });
    restoreConfig();
  });

  test('push campaign: channel, category, player filter, live count, sandbox delivery, opened and the dead token', async ({ page }) => {
    await open(page, '#/publication/notifications/new?tagId=' + tagId);
    const r = root(page);
    await r.locator('button[role="radio"][data-value="push_in_app"]').click();
    await expect(r).toContainText('پیام روی دستگاه‌های ثبت‌شدهٔ بازیکن می‌رود');
    await fillTranslated(field(r, /^عنوان$/), '', { fa: 'حراج {player} ' + k });
    await fillTranslated(field(r, /^متن$/), '', { fa: 'امروز سکه‌ها نصف قیمت‌اند' });
    await r.locator('[data-nt-name]').fill('کمپین آزمایشی ' + k);
    await r.locator('[data-nt-category-pick] button[aria-haspopup="listbox"]').click();
    await r.locator('[data-nt-category-pick] .eco-pop input').fill('تخفیف آزمایشی ' + k);
    await r.locator('[data-nt-category-pick] .eco-pop [role="option"]', { hasText: 'تخفیف آزمایشی ' + k }).first().click();
    await expect(r.locator('[data-recipient-count]')).toHaveAttribute('data-recipient-count', '3');
    await expect(r.locator('[data-nt-opted-out]')).toHaveAttribute('data-nt-opted-out', '1');
    await r.locator('[data-nt-audience-box] summary').click();
    await field(r.locator('[data-nt-audience-box]'), 'سکوها').locator('label', { hasText: 'وب' }).click();
    await expect(r.locator('[data-recipient-count]')).toHaveAttribute('data-recipient-count', '2');
    await expect(r.locator('[data-nt-with-push]')).toHaveAttribute('data-nt-with-push', '1');
    await r.locator('[data-nt-send]').click();
    await expect(page).toHaveURL(/#\/publication\/notifications\/[0-9a-f-]{36}$/);
    const id = page.url().split('/').pop() as string;
    await expect(r.locator('[data-nt-status]')).toHaveAttribute('data-nt-status', 'done', { timeout: 30_000 });
    await expect(r.locator('[data-nt-kind="campaign"]')).toBeVisible();
    await expect(r.locator('[data-nt-audience]')).toContainText('وب');
    const n = await pub.get('/notifications/' + id);
    expect(n).toMatchObject({ kind: 'campaign', channel: 'push_in_app', recipientCount: 2, sentCount: 1, skippedCount: 1, pushSent: 1, name: 'کمپین آزمایشی ' + k });
    await expect(r.locator('[data-recipient-error="opted_out"]')).toHaveText('این دسته را خاموش کرده');
    const inbox = await rpc(A, 'publication_inbox', { lang: 'fa' });
    const item = inbox.data.items.find((x: any) => x.id === id);
    expect(item.title).toBe('حراج ' + A.username + ' ' + k);
    expect(item.category.key).toBe(catKey);
    expect((await rpc(A, 'publication_notification_opened', { notificationId: id })).data.counted).toBe(true);
    expect((await rpc(A, 'publication_notification_opened', { notificationId: id })).data.counted).toBe(false);
    await page.reload();
    await expect(r.locator('[data-recipient-opened]')).toHaveCount(1);
    expect((await pub.get('/notifications/' + id)).openedCount).toBe(1);

    const dead = await pub.post('/notifications', { channel: 'push', title: { fa: 'مرده ' + k }, body: { fa: 'x' }, categoryId: catId, target: { mode: 'users', userIds: [B.id] } });
    await expect.poll(async () => (await pub.get('/notifications/' + dead.id)).status, { timeout: 30_000 }).toBe('failed');
    expect(sql(`SELECT is_active FROM publication_push_device WHERE user_id='${B.id}'`)).toBe('f');
    const devs = await pub.get('/push-devices?userId=' + B.id);
    expect(devs.items[0]).toMatchObject({ isActive: false, lastError: expect.stringContaining('UNREGISTERED') });
  });

  test('scheduled campaign in the project zone, cancel, and the test send', async ({ page }) => {
    await open(page, '#/publication/notifications/new');
    const r = root(page);
    await r.locator('button[role="radio"][data-value="push"]').click();
    await fillTranslated(field(r, /^عنوان$/), '', { fa: 'فردا ' + k });
    await fillTranslated(field(r, /^متن$/), '', { fa: 'فردا بیا' });
    await r.locator('[data-nt-category-pick] button[aria-haspopup="listbox"]').click();
    await r.locator('[data-nt-category-pick] .eco-pop input').fill('تخفیف آزمایشی ' + k);
    await r.locator('[data-nt-category-pick] .eco-pop [role="option"]', { hasText: 'تخفیف آزمایشی ' + k }).first().click();
    await r.locator('[data-nt-test] button[aria-haspopup="listbox"]').click();
    await r.locator('[data-nt-test] .eco-pop input').fill(A.username);
    await r.locator('[data-nt-test] .eco-pop [role="option"]').first().click();
    await r.locator('[data-nt-test-run]').click();
    await expect(r.locator('[data-nt-test-done]')).toBeVisible();
    const testId = await r.locator('[data-nt-test-done]').getAttribute('data-nt-test-done');
    await expect.poll(async () => (await pub.get('/notifications/' + testId)).status, { timeout: 30_000 }).toBe('done');
    expect((await pub.get('/notifications/' + testId)).kind).toBe('test');
    await r.locator('button[role="radio"][data-value="all"]').click();
    await r.locator('[data-nt-when] button', { hasText: 'در زمان مشخص' }).click();
    const day = new Date(Date.now() + 30 * 86400_000).toISOString().slice(0, 10);
    await r.locator('[data-zone-input]').fill(day + 'T09:30');
    await r.locator('[data-nt-send]').click();
    await page.locator('[role="alertdialog"] button', { hasText: 'زمان‌بندی کن' }).click();
    await expect(page).toHaveURL(/#\/publication\/notifications\/[0-9a-f-]{36}$/);
    const id = page.url().split('/').pop() as string;
    await expect(r.locator('[data-nt-status]')).toHaveAttribute('data-nt-status', 'scheduled');
    const n = await pub.get('/notifications/' + id);
    expect(new Date(n.scheduledAt).toISOString()).toBe(day + 'T06:00:00.000Z');
    expect(n.recipientCount).toBe(0);
    await r.locator('[data-nt-cancel]').click();
    await page.locator('[role="alertdialog"] button', { hasText: 'لغو کن' }).click();
    await expect(r.locator('[data-nt-status]')).toHaveAttribute('data-nt-status', 'cancelled');
    await expect(toasts(page).last()).toContainText('ارسال لغو شد');
    const again = await consoleApi('POST', '/v2/console/publication/notifications/' + id + '/cancel');
    expect(again.status).toBe(409);
    await go(page, '#/publication/notifications?kind=campaign&status=cancelled');
    await expect(rows(page).filter({ hasText: 'فردا ' + k })).toHaveCount(1);
  });

  test('automatic rule: editor with the live sentence, one message per event, the repeat gap, opt-out, stats and the test send', async ({ page }) => {
    const type = await eco.post('/event-types', { key: evKey, name: { fa: 'رویداد پوش ' + k }, clientAllowed: true, isActive: true, rateLimit: 30 });
    made.types.push(type.id);
    await open(page, '#/publication/auto-rules');
    await expect(rows(page).filter({ hasText: 'star_auto_badge' })).toHaveCount(1);
    await go(page, '#/publication/auto-rules/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(ruleKey);
    await pickRef(p, /^رویداد$/, evKey);
    await expect(p.locator('[data-ar-sentence]')).toContainText('وقتی رویداد «' + evKey + '» برای بازیکن ثبت شود');
    await pickRef(p, 'دستهٔ اعلان', 'تخفیف آزمایشی ' + k);
    await fillTranslated(p, 'title', { fa: '{player} امتیاز گرفت' });
    await fillTranslated(p, 'body', { fa: 'امتیاز تو {value} شد' });
    await field(p, 'فاصلهٔ تکرار').locator('input').fill('60');
    await expect(p.locator('[data-ar-sentence]')).toContainText('هر بازیکن هر ۱ ساعت یک بار');
    await expect(p.locator('[data-ar-vars]')).toContainText('{value}');
    await savePanel(page);
    const rule = await byKey('auto-rules', ruleKey);
    expect(rule).toMatchObject({ trigger: 'event', eventKey: evKey, channel: 'push_in_app', cooldownMinutes: 60 });
    const half = await consoleApi('PUT', '/v2/console/publication/auto-rules/' + rule.id, { ...rule, quietStart: '22:00', quietEnd: '' });
    expect(half.status).toBe(400);
    expect(half.data.error).toBe('INVALID_QUIET_HOURS');

    for (const who of [A, A, C]) expect((await rpc(who, 'economy_event_record', { key: evKey, value: 7 })).status).toBe(200);
    await expect.poll(() => Number(sql(`SELECT count(*) FROM publication_auto_fire WHERE rule_id='${rule.id}' AND status <> 'pending'`)), { timeout: 45_000 }).toBe(3);
    expect(sql(`SELECT string_agg(status || ':' || reason, ',' ORDER BY status, reason) FROM publication_auto_fire WHERE rule_id='${rule.id}'`)).toBe('sent:,skipped:cooldown,skipped:opted_out');
    await expect.poll(async () => (await rpc(A, 'publication_inbox', { lang: 'fa' })).data.items.some((x: any) => x.title === A.username + ' امتیاز گرفت' && x.body === 'امتیاز تو 7 شد'), { timeout: 30_000 }).toBe(true);

    await go(page, '#/publication/auto-rules/' + rule.id);
    await settlePanel(page);
    const st = panel(page).locator('[data-ar-stats]');
    await expect(st.locator('[data-ar-sent]')).toContainText('۱');
    await expect(st.locator('[data-ar-skipped]')).toContainText('۲');
    await expect(st.locator('[data-ar-fire="skipped"][data-reason="opted_out"]')).toContainText('دسته را خاموش کرده');
    await expect(st.locator('[data-ar-fire="skipped"][data-reason="cooldown"]')).toContainText('فاصلهٔ تکرار');
    await st.locator('[data-ar-test] button[aria-haspopup="listbox"]').click();
    await st.locator('[data-ar-test] .eco-pop input').fill(C.username);
    await st.locator('[data-ar-test] .eco-pop [role="option"]').first().click();
    await st.locator('[data-ar-test-run]').click();
    await expect(st.locator('[data-ar-test-done]')).toBeVisible();
    await page.keyboard.press('Escape');
    await search(page, ruleKey);
    await expect(rows(page).first().locator('[data-ar-row-stats]')).toContainText('۱ رفته در ۷ روز');
    const audit = sql(`SELECT string_agg(DISTINCT action, ',' ORDER BY action) FROM console_audit_log WHERE message LIKE 'publication % auto-rules%' AND create_time > now() - interval '10 minutes'`);
    expect(audit.split(',')).toEqual(expect.arrayContaining(['CREATE']));
  });

  test('sidebar, page search, English and 390px in dark mode without sideways scroll', async ({ page, spawn }) => {
    await open(page, '#/publication/auto-rules');
    await expect(page.locator('.pl-side a.pl-nav[href^="#/publication/auto-rules"] svg')).toHaveCount(1);
    await expect(page.locator('.pl-side a.pl-nav[href^="#/publication/notify-categories"]')).toContainText('دسته‌های اعلان');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('اعلان‌های خودکار');
    await expect(page.locator('.pl-pal-row').first()).toContainText('اعلان‌های خودکار');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/publication\/auto-rules/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/publication/auto-rules?q=star_auto_wheel');
    await expect(rows(page).first().locator('[data-ar-when]')).toContainText('When the free turn of "star_wheel_daily" is ready');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);

    const m = await spawn({ signedIn: true, mode: 'dark', mobile: true });
    for (const hash of ['#/publication/auto-rules', '#/publication/notify-categories', '#/publication/notifications/new']) {
      await open(m, hash);
      await expect(m.locator('html')).toHaveClass(/dark/);
      await noHorizontalScroll(m);
    }
  });
});
