import fs from 'node:fs';
import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, root, settle, toasts } from '../../lib/console';
import { base, consoleApi } from '../../lib/api';
import { bearerOf, dropConsoleUser, makeConsoleUser, signIn } from '../../lib/support';
import { CONSOLE_URL, SERVER_KEY } from '../../lib/env';

const SECRET = /[0-9a-f]{64}|NAKAMA_CONSOLE_PASSWORD=|http_key\s*[:=]\s*\S|RkFront-Test/;

function withoutKey(html: string) {
  return html.split(SERVER_KEY).join('');
}

async function rules() {
  return base.get('/server-rules');
}

async function setHandoff(h: { enabled: boolean; audience: string; showServerKey?: boolean }) {
  const handoff = { showServerKey: true, ...h };
  const cur = await rules();
  delete cur.updatedAt;
  delete cur.updatedBy;
  delete cur.pendingDeletions;
  await base.put('/server-rules', { ...cur, handoff });
}

test.describe.configure({ mode: 'serial' });

test.describe('front-end handoff', () => {
  test.afterAll(async () => {
    await setHandoff({ enabled: true, audience: 'all' });
  });

  test('the console API serves meta and one file; without a login it refuses', async () => {
    expect((await fetch(CONSOLE_URL + '/v2/console/base/handoff/download')).status).toBe(401);
    expect((await fetch(CONSOLE_URL + '/v2/console/base/handoff')).status).toBe(401);
    const meta = await base.get('/handoff');
    expect(meta.visible).toBe(true);
    expect(meta.settings).toEqual({ enabled: true, audience: 'all', showServerKey: true });
    const general = meta.changelog.filter((e: any) => e.scope === 'general');
    expect(general.length).toBeGreaterThan(0);
    const dates = meta.changelog.map((e: any) => e.date);
    expect([...dates].sort().reverse()).toEqual(dates);
    const res = await consoleApi('GET', '/v2/console/base/handoff/download');
    expect(res.status).toBe(200);
    const html = String(res.data);
    expect(html).toContain('<html lang="fa" dir="rtl">');
    expect(html).toContain('id="part-general"');
    expect(html).toContain('@font-face{font-family:Vazirmatn');
    expect(html).not.toMatch(/<link[^>]+href="http|<script[^>]+src=/);
    expect(withoutKey(html)).not.toMatch(SECRET);
    if (meta.project) {
      expect(html).toContain('id="part-project"');
      expect(meta.changelog.some((e: any) => e.scope === 'project')).toBe(true);
      expect(html).toContain('<span class="badge project">اختصاصی</span>');
    } else {
      expect(html).toContain('data-no-project');
    }
    expect(html).toContain('<span class="badge general">عمومی</span>');
    for (const e of meta.changelog) expect(html).toContain(e.title.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&#34;').replace(/'/g, '&#39;'));
  });

  test('the header button opens the panel; the file downloads with both parts and the merged list', async ({ page }) => {
    const meta = await base.get('/handoff');
    await open(page, '#/economy/overview');
    const btn = page.locator('button[data-pl-handoff]');
    await expect(btn).toBeVisible();
    await expect(btn).toHaveText('راهنمای فرانت');
    await expect(btn).toHaveAttribute('title', 'دانلود راهنمای تیم فرانت و فهرست تازه‌ها');
    await btn.click();
    const panel = page.locator('[data-handoff] [role="dialog"]');
    await expect(panel).toBeVisible();
    await expect(panel).toHaveAttribute('aria-label', 'راهنمای فرانت');
    const body = panel.locator('[data-handoff-panel]');
    await expect(body).toBeVisible();
    await expect(page.locator('[data-handoff-download]')).toBeFocused();
    await expect(body.locator('[data-handoff-part="general"]')).toContainText('بخش عمومی');
    if (meta.project) {
      await expect(body.locator('[data-handoff-part="project"]')).toContainText(meta.project.title);
      await expect(body.locator('[data-handoff-no-project]')).toHaveCount(0);
    } else {
      await expect(body.locator('[data-handoff-part="project"]')).toHaveCount(0);
      await expect(body.locator('[data-handoff-no-project]')).toContainText('هنوز بخش اختصاصی ندارد');
    }
    const rows = body.locator('[data-handoff-changes] > li[data-scope]');
    await expect(rows).toHaveCount(meta.changelog.length);
    await expect(rows.first()).toContainText(meta.changelog[0].title);
    await expect(rows.first().locator('time')).toHaveText(/^[۰-۹]/);
    await body.locator('[data-handoff-filter] [data-value="general"]').click();
    await expect(rows).toHaveCount(meta.changelog.filter((e: any) => e.scope === 'general').length);
    await expect(body.locator('[data-handoff-changes] > li[data-scope="project"]')).toHaveCount(0);
    if (meta.project) {
      await body.locator('[data-handoff-filter] [data-value="project"]').click();
      await expect(rows).toHaveCount(meta.changelog.filter((e: any) => e.scope === 'project').length);
    } else {
      await expect(body.locator('[data-handoff-filter] [data-value="project"]')).toHaveCount(0);
    }
    await body.locator('[data-handoff-filter] [data-value="all"]').click();
    const wait = page.waitForEvent('download');
    await page.locator('[data-handoff-download]').click();
    const dl = await wait;
    expect(dl.suggestedFilename()).toBe(meta.fileName);
    const html = fs.readFileSync(await dl.path(), 'utf8');
    expect(html).toContain('id="part-general"');
    expect(html).toContain('id="changes"');
    expect(withoutKey(html)).not.toMatch(SECRET);
    if (meta.project) expect(html).toContain('id="part-project"');
    const changeRows = html.match(/<tr data-scope="(general|project)">/g) || [];
    expect(changeRows.length).toBe(meta.changelog.length);
    await expect(toasts(page).filter({ hasText: 'فایل راهنما دانلود شد.' }).first()).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('[data-handoff]')).toHaveCount(0);
    await expect(btn).toBeFocused();
  });

  test('the live server key is in the panel, the file and the SDK page; off hides it everywhere', async ({ page }) => {
    const meta = await base.get('/handoff');
    expect(meta.connection.keyShown).toBe(true);
    expect(meta.connection.serverKey).toBe(SERVER_KEY);
    expect(meta.connection.defaultKey).toBe(SERVER_KEY === 'defaultkey');
    const res = await consoleApi('GET', '/v2/console/base/handoff/download');
    const html = String(res.data);
    expect(html).not.toContain('{{SERVER_');
    expect(html).toContain('<section class="step" id="connection" data-connection data-key-shown="true"');
    expect(html).toContain('<td data-server-key>');
    expect(html).toContain('new Client(&#39;' + SERVER_KEY + '&#39;, ');
    expect(html).toContain('<li><a href="#connection">');
    if (SERVER_KEY === 'defaultkey') expect(html).toContain('data-default-key');
    else expect(html).not.toContain('data-default-key');
    if (meta.connection.hostSet) expect(html).toContain(meta.connection.url);
    else expect(html).toContain('data-host-missing');

    await open(page, '#/economy/overview');
    await page.locator('button[data-pl-handoff]').click();
    const box = page.locator('[data-handoff-panel] [data-connection]');
    await expect(box).toHaveAttribute('data-key-shown', 'true');
    await expect(box.locator('[data-conn="key"] code')).toHaveText(SERVER_KEY);
    await expect(box.locator('[data-conn="key"] [data-conn-copy]')).toBeVisible();
    await page.keyboard.press('Escape');

    await open(page, '#/system/sdk');
    const sdk = root(page).locator('[data-sdk-connection] [data-connection]');
    await expect(sdk.locator('[data-conn="key"] code')).toHaveText(SERVER_KEY);
    await expect(root(page).locator('[data-sdk-page]')).not.toContainText('از مدیر سرور بگیرید');
    await page.context().grantPermissions(['clipboard-read', 'clipboard-write']).catch(() => null);
    await sdk.locator('[data-conn="key"] [data-conn-copy]').click();
    await expect(toasts(page).first()).toBeVisible();

    await setHandoff({ enabled: true, audience: 'all', showServerKey: false });
    try {
      const off = await base.get('/handoff');
      expect(off.settings.showServerKey).toBe(false);
      expect(off.connection.keyShown).toBe(false);
      expect(off.connection.serverKey).toBeUndefined();
      const offHtml = String((await consoleApi('GET', '/v2/console/base/handoff/download')).data);
      expect(offHtml).not.toContain(SERVER_KEY);
      expect(offHtml).toContain('data-key-hidden');
      expect(offHtml).toContain('ASK_THE_SERVER_ADMIN');
      expect(offHtml).not.toContain('{{SERVER_');
      expect(offHtml).not.toMatch(SECRET);
      const conn = await base.get('/handoff/connection');
      expect(conn.connection.serverKey).toBeUndefined();
      await open(page, '#/economy/overview');
      await open(page, '#/system/sdk');
      const hidden = root(page).locator('[data-sdk-connection] [data-connection]');
      await expect(hidden).toHaveAttribute('data-key-shown', 'false');
      await expect(hidden.locator('[data-conn="key"]')).toContainText('از مدیر سرور بگیرید');
      await expect(hidden.locator('[data-conn="key"] [data-conn-copy]')).toHaveCount(0);
      await expect(root(page).locator('[data-sdk-page]')).not.toContainText(SERVER_KEY);
      await open(page, '#/system/server-rules');
      await expect(page.locator('[data-settings-form="server-rules"]')).toContainText('نمایش کلید سرور در راهنما و صفحهٔ SDK');
    } finally {
      await setHandoff({ enabled: true, audience: 'all' });
    }
  });

  test('the downloaded file opens offline: RTL, light and dark, no sideways scroll on a phone', async ({ spawn }) => {
    const res = await consoleApi('GET', '/v2/console/base/handoff/download');
    const p = await spawn({ signedIn: false, mobile: true, width: 390, height: 844, mode: 'dark' });
    await p.route('**/*', (r) => (r.request().url().startsWith('data:') ? r.continue() : r.abort()));
    await p.emulateMedia({ colorScheme: 'dark' });
    await p.setContent(String(res.data));
    await expect(p.locator('html')).toHaveAttribute('dir', 'rtl');
    await expect(p.locator('#changes tbody tr').first()).toBeVisible();
    const dark = await p.evaluate(() => getComputedStyle(document.body).backgroundColor);
    await p.locator('[data-theme-toggle]').click();
    await expect(p.locator('html')).toHaveAttribute('data-theme', 'light');
    expect(await p.evaluate(() => getComputedStyle(document.body).backgroundColor)).not.toBe(dark);
    expect(await p.evaluate(() => document.fonts.check('16px Vazirmatn'))).toBe(true);
    await noHorizontalScroll(p);
  });

  test('settings: admins only hides it from other console users; off hides it for everyone', async ({ page, spawn }) => {
    await open(page, '#/system/server-rules');
    const form = page.locator('[data-settings-form="server-rules"]');
    await expect(form).toContainText('راهنمای فرانت');
    await expect(form).toContainText('همهٔ کاربران کنسول');
    await setHandoff({ enabled: true, audience: 'admins' });
    const user = await makeConsoleUser('e2ehf' + Date.now().toString(36), { ACCOUNT_NOTES: { read: true, write: true } });
    try {
      const p = await spawn({ signedIn: false });
      await signIn(p, user);
      await settle(p);
      const bearer = await bearerOf(p);
      const meta = await consoleApi('GET', '/v2/console/base/handoff', undefined, bearer);
      expect(meta.status).toBe(200);
      expect(meta.data).toMatchObject({ visible: false, reason: 'HANDOFF_ADMINS_ONLY' });
      expect(meta.data.changelog).toBeUndefined();
      expect((await consoleApi('GET', '/v2/console/base/handoff/download', undefined, bearer)).status).toBe(403);
      const conn = await consoleApi('GET', '/v2/console/base/handoff/connection', undefined, bearer);
      expect(conn.status).toBe(200);
      expect(conn.data.connection.keyShown).toBe(false);
      expect(JSON.stringify(conn.data)).not.toContain(SERVER_KEY);
      await expect(p.locator('button[data-pl-handoff]')).toBeHidden();
      expect((await consoleApi('PUT', '/v2/console/base/server-rules', {}, bearer)).status).toBe(403);
      await setHandoff({ enabled: true, audience: 'all' });
      expect((await consoleApi('GET', '/v2/console/base/handoff', undefined, bearer)).data.visible).toBe(true);
      expect((await consoleApi('GET', '/v2/console/base/handoff/download', undefined, bearer)).status).toBe(200);
    } finally {
      await dropConsoleUser(user.user);
    }
    await setHandoff({ enabled: false, audience: 'all' });
    expect((await consoleApi('GET', '/v2/console/base/handoff/download')).status).toBe(403);
    await open(page, '#/economy/overview');
    await page.reload();
    await settle(page);
    await expect(page.locator('button[data-pl-handoff]')).toBeHidden();
    const bad = await consoleApi('PUT', '/v2/console/base/server-rules', { ...(await rules()), handoff: { enabled: true, audience: 'everyone' } });
    expect(bad.status).toBe(400);
    await setHandoff({ enabled: true, audience: 'all' });
    await page.reload();
    await settle(page);
    await expect(page.locator('button[data-pl-handoff]')).toBeVisible();
  });

  test('English, dark, 390px: icon-only button, English panel, no sideways scroll', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/economy/overview');
    const btn = m.locator('button[data-pl-handoff]');
    await expect(btn).toBeVisible();
    await expect(btn.locator('.pl-sdk-label')).toBeHidden();
    await expect(btn).toHaveAttribute('aria-label', 'Download the front-end team handoff and see what is new');
    await btn.click();
    const body = m.locator('[data-handoff-panel]');
    await expect(body).toBeVisible();
    await expect(m.locator('[data-handoff-download]')).toHaveText('Download the handoff file');
    await expect(body).toContainText("What's new");
    await expect(body.locator('[data-handoff-changes] time').first()).toHaveText(/^[A-Z][a-z]{2} \d/);
    await noHorizontalScroll(m);
  });
});
