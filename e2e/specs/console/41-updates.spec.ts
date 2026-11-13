import fs from 'node:fs';
import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi } from '../../lib/api';
import { sql } from '../../lib/db';
import { FAKE_GITHUB } from '../../lib/env';
import { noHorizontalScroll, open, openSettings, root, settle, toasts } from '../../lib/console';
import type { Page, TestInfo } from '@playwright/test';

const TOKEN = 'github_pat_e2e' + 'A'.repeat(30);
const SHOTS = process.env.E2E_SHOTS_DIR || '';

test.describe.configure({ mode: 'serial' });

async function fake(body: Record<string, unknown>) {
  const r = await fetch(FAKE_GITHUB + '/__fake/scenario', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  expect(r.status).toBe(200);
  return (await r.json()) as { scenario: any; calls: Record<string, number> };
}

async function calls() {
  const r = await fetch(FAKE_GITHUB + '/__fake/scenario');
  return ((await r.json()) as { calls: Record<string, number> }).calls;
}

function freeChecks() {
  sql(`DELETE FROM cluster_rate_limit WHERE key = 'base:updates:check'`);
}

async function settled(pred: (v: any) => boolean = (v) => v.state.status === 'ok' && !v.state.queued && v.current) {
  let last: any = null;
  await expect.poll(async () => { last = await base.get('/updates'); return pred(last); }, { timeout: 100_000, intervals: [500, 1000] }).toBe(true);
  return last;
}

async function queue() {
  const r = await consoleApi('POST', '/v2/console/base/updates/check', {});
  expect(r.status, JSON.stringify(r.data).slice(0, 200)).toBe(202);
  return r.data;
}

async function fresh(body: Record<string, unknown> = {}) {
  await fake(body);
  freeChecks();
  sql(`UPDATE base_updates_state SET next_at = now() WHERE id = 1`);
  await queue();
  return settled();
}

async function home(page: Page) {
  await page.goto('/#/');
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.locator('#main .pl-attn').waitFor({ timeout: 20_000 });
}

async function shot(page: Page, info: TestInfo, name: string) {
  await page.evaluate(() => { document.getElementById('main')?.scrollTo(0, 0); window.scrollTo(0, 0); });
  await page.locator('.pl-rail .pl-mode[data-pl-mode="system"]').first().waitFor({ state: 'attached', timeout: 15_000 }).catch(() => null);
  await page.waitForTimeout(400);
  const body = await page.screenshot({ fullPage: false });
  await info.attach(name, { body, contentType: 'image/png' });
  if (SHOTS) fs.writeFileSync(path.join(SHOTS, name + '.png'), body);
}

test.beforeAll(async () => {
  await fake({ reset: true });
  sql(`DELETE FROM base_updates_settings`);
  await fresh();
});

test.afterAll(async () => {
  sql(`DELETE FROM base_updates_settings`);
  await fresh({ reset: true });
  freeChecks();
});

test.describe('updates: page', () => {
  test('version cards, release notes first, warnings, commits, files by area and the copy-only guide', async ({ page }, info) => {
    const v = await base.get('/updates');
    expect(v.result.aheadBy).toBe(4);
    expect(v.build.parent.repo).toBe('heroiclabs/nakama');
    await open(page, '#/system/updates');
    const r = root(page);
    await expect(r.locator('h1')).toHaveText('به‌روزرسانی‌ها');
    const mine = r.locator('[data-up-this]');
    await expect(mine).toContainText('این سرور');
    await expect(mine).toContainText(v.build.branch);
    await expect(mine).toContainText(v.build.parent.commit.slice(0, 7));
    await expect(mine).toContainText('۴');
    await expect(mine).toContainText('کامیت عقب‌تر از ریپوی بالایی');
    const parent = r.locator('[data-up-parent]');
    await expect(parent).toContainText('heroiclabs/nakama · master');
    await expect(parent).toContainText('بررسی شد');
    await expect(parent).toContainText('v3.43.0');
    await expect(parent).toContainText('هر ۶ ساعت');
    await expect(parent).toContainText('کامیت تازه');
    await expect(parent.locator('a[href*="github.com/heroiclabs/nakama/compare/"]')).toHaveCount(1);

    const order = await r.locator('[data-up-warnings], [data-up-releases], [data-up-commits], [data-up-files], [data-up-guide]').evaluateAll((els) => els.map((e) => [...e.attributes].map((a) => a.name).find((n) => n.startsWith('data-up-'))));
    expect(order).toEqual(['data-up-warnings', 'data-up-releases', 'data-up-commits', 'data-up-files', 'data-up-guide']);

    await expect(r.locator('[data-up-warning="migrations"]')).toContainText('20261201000000-session-cache.sql');
    await expect(r.locator('[data-up-warning="migrations"]')).toContainText('۱ مایگریشن تازه');
    await expect(r.locator('[data-up-warning="env"]')).toContainText('NAKAMA_SESSION_WARMUP');
    await expect(r.locator('[data-up-warning="goVersion"]')).toContainText('go 1.27.3 → 1.27.4');
    await expect(r.locator('[data-up-warning="deps"]')).toContainText('github.com/jackc/pgx/v5');
    await expect(r.locator('[data-up-warning="breaking"]')).toContainText('۲ تغییر ناسازگار');

    const rel = r.locator('[data-up-release="v3.43.0"]');
    await expect(rel).toContainText('Session cache warm-up flag.');
    await expect(rel).toContainText('ناسازگار');
    await expect(r.locator('[data-up-release="v3.42.0"]')).toHaveCount(0);
    await expect(r.locator('[data-up-tags] a')).toHaveText(['v3.43.0']);

    const commits = r.locator('[data-up-commit]');
    await expect(commits).toHaveCount(4);
    await expect(commits.first()).toContainText('Release v3.43.0');
    await expect(commits.nth(1)).toContainText('ناسازگار');
    await expect(commits.first().locator('a')).toHaveAttribute('href', /^https:\/\/github\.com\/heroiclabs\/nakama\/commit\/[0-9a-f]{40}$/);
    await expect(r.locator('[data-up-commits]')).toContainText('۴ کامیت');

    const areas = await r.locator('[data-up-area]').evaluateAll((els) => els.map((e) => e.getAttribute('data-up-area')));
    expect(areas).toEqual(['migrations', 'core', 'console', 'deps', 'build', 'e2e', 'docs']);
    await expect(r.locator('[data-up-area="migrations"]')).toHaveAttribute('open', '');
    await expect(r.locator('[data-up-area="migrations"]')).toContainText('migrate/sql/20261201000000-session-cache.sql');
    await r.locator('[data-up-area="deps"] summary').click();
    await expect(r.locator('[data-up-area="deps"]')).toContainText('vendor/modules.txt');
    await expect(r.locator('[data-up-area="deps"] summary')).toContainText('۳ فایل');
    await expect(r.locator('[data-up-area="console"] summary')).toContainText('کنسول');
    await expect(r.locator('[data-up-files]')).toContainText('۱۰ فایل');

    const guide = r.locator('[data-up-guide]');
    await expect(guide).toContainText('هیچ‌کدام را اجرا نمی‌کند');
    await expect(guide).toContainText('git fetch upstream master');
    await expect(guide).toContainText('git merge upstream/master');
    await expect(guide).toContainText('build/upstream-sync.sh record upstream master');
    await expect(guide).toContainText('migrate up');
    await expect(guide.locator('pre')).toHaveCount(6);
    await expect(guide).toContainText('copilot-rast/nakama-farsi');
    await guide.locator('[data-up-copy]').nth(1).click();
    await expect(toasts(page).last()).toContainText('کپی شد');
    await page.evaluate(() => window.scrollTo(0, 0));
    await shot(page, info, 'updates-fa-light');
  });

  test('check again queues a check, shows the new count, and a repeat check is answered from the ETag', async ({ page }) => {
    await fake({ ahead: 2 });
    freeChecks();
    await open(page, '#/system/updates');
    const parent = root(page).locator('[data-up-parent]');
    await expect(parent).toContainText('۴');
    await parent.locator('[data-up-check]').click();
    await expect(toasts(page).last()).toContainText('بررسی در صف قرار گرفت');
    await expect(parent).toContainText('۲', { timeout: 30_000 });
    await expect(root(page).locator('[data-up-commit]')).toHaveCount(2);
    await expect(root(page).locator('[data-up-this]')).toContainText('۲');
    const v = await settled();
    expect(v.result.aheadBy).toBe(2);
    expect(v.state.requestedAt).toBeTruthy();

    const before = await calls();
    await queue();
    await settled((x) => x.state.status === 'ok' && !x.state.queued && x.state.checks > v.state.checks);
    const after = await calls();
    expect(after.compare).toBe(before.compare + 1);
    expect(after.notModified).toBe(before.notModified + 3);
    expect((await base.get('/updates')).result.aheadBy).toBe(2);
    const res = await consoleApi('GET', '/v2/console/base/updates/summary');
    expect(res.status).toBe(200);
    expect(res.data.newCommits).toBe(2);
  });

  test('a GitHub rate limit keeps the last result and says when the next check runs; an error is shown too', async ({ page }) => {
    await fake({ status: 429, retryAfter: 900 });
    freeChecks();
    await queue();
    await settled((x) => x.state.status === 'rate_limited');
    await open(page, '#/system/updates');
    const n = root(page).locator('[data-up-notice="rate_limited"]');
    await expect(n).toContainText('سهم درخواست گیت‌هاب تمام شده');
    await expect(root(page).locator('[data-up-parent]')).toContainText('محدودیت گیت‌هاب');
    await expect(root(page).locator('[data-up-commit]').first()).toBeVisible();
    const again = await consoleApi('POST', '/v2/console/base/updates/check', {});
    expect(again.status).toBe(202);
    await page.waitForTimeout(6000);
    expect((await base.get('/updates')).state.status).toBe('rate_limited');

    await fake({ status: 500 });
    sql(`UPDATE base_updates_state SET next_at = now() WHERE id = 1`);
    await settled((x) => x.state.status === 'error');
    await page.reload();
    await root(page).locator('[data-up-parent]').waitFor();
    await expect(root(page).locator('[data-up-notice="error"]')).toContainText('500');
    await expect(root(page).locator('[data-up-commit]').first()).toBeVisible();
    await fresh({ status: 0, ahead: 4 });
  });

  test('badges: sidebar count, top bar pill, dashboard needs-attention row and Ctrl K', async ({ page }) => {
    await home(page);
    const pill = page.locator('[data-pl-updates]');
    await expect(pill).toBeVisible({ timeout: 20_000 });
    await expect(pill).toContainText('۴');
    await expect(pill).toHaveAttribute('aria-label', '۴ تغییر تازه در ریپوی بالایی');
    const row = page.locator('#main .pl-attn a.pl-attn-row[href="#/system/updates"]');
    await expect(row).toContainText('به‌روزرسانی‌ها');
    await expect(row).toContainText('۴');
    await page.locator('.pl-rail .pl-mode[data-pl-mode="system"]').click();
    await expect(page).toHaveURL(/#\/system\/updates/);
    await expect(page.locator('.pl-side a.pl-nav[href="#/system/updates"]')).toContainText('۴');
    await expect(page.locator('.x-drawer a[data-economy-page="updates"] .eco-nav-badge').first()).toHaveText('۴');
    await expect(page.locator('.pl-crumbs .pl-crumb-current')).toHaveText('به‌روزرسانی‌ها');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('به‌روز');
    await expect(page.locator('.pl-pal-row').first()).toContainText('به‌روزرسانی‌ها');
    await page.locator('.pl-pal-input').fill('ریپوی');
    await expect(page.locator('.pl-pal-row').first()).toContainText('ریپوی بالایی');
    await page.keyboard.press('Escape');
    await home(page);
    await pill.click();
    await expect(page).toHaveURL(/#\/system\/updates/);
  });

  test('EN in dark mode uses Latin digits and English labels', async ({ spawn }, info) => {
    const page = await spawn({ signedIn: true, locale: 'en', mode: 'dark' });
    await open(page, '#/system/updates');
    const r = root(page);
    await expect(r.locator('h1')).toHaveText('Updates');
    await expect(r.locator('[data-up-parent]')).toContainText('every 6 h');
    await expect(r.locator('[data-up-commits]')).toContainText('4 commits');
    await expect(r.locator('[data-up-warning="migrations"]')).toContainText('1 new migrations');
    await expect(r).toContainText('How to update');
    await expect(page.locator('[data-pl-updates]')).toContainText('4');
    await shot(page, info, 'updates-en-dark');
    const dark = await spawn({ signedIn: true, mode: 'dark' });
    await open(dark, '#/system/updates');
    await expect(root(dark).locator('[data-up-parent]')).toContainText('۴');
    await shot(dark, info, 'updates-fa-dark');
  });

  test('390px: cards stack, nothing scrolls sideways, the top bar keeps only the count', async ({ spawn }, info) => {
    const page = await spawn({ signedIn: true, mobile: true });
    await open(page, '#/system/updates');
    await expect(root(page).locator('[data-up-guide]')).toBeVisible();
    await noHorizontalScroll(page);
    const a = await root(page).locator('[data-up-this]').boundingBox();
    const b = await root(page).locator('[data-up-parent]').boundingBox();
    expect(b!.y).toBeGreaterThan(a!.y + a!.height - 1);
    expect(a!.width).toBeLessThanOrEqual(390);
    await expect(page.locator('[data-pl-updates] .pl-up-label')).toBeHidden();
    await expect(page.locator('[data-pl-updates] .pl-up-n')).toHaveText('۴');
    await shot(page, info, 'updates-mobile');
    await root(page).locator('[data-up-guide]').scrollIntoViewIfNeeded();
    await noHorizontalScroll(page);
  });
});

test.describe('updates: parent repo settings', () => {
  test('defaults, field errors, fork branches, a sealed masked token, the audit log and the base mismatch notice', async ({ page }, info) => {
    await openSettings(page, 'updates-source');
    const form = root(page).locator('[data-settings-form="updates-source"]');
    const repo = form.getByRole('textbox', { name: 'نشانی ریپو' });
    const branch = form.getByRole('combobox', { name: 'شاخه' });
    const token = form.locator('[data-up-token]');
    await expect(repo).toHaveValue('heroiclabs/nakama');
    await expect(branch).toHaveValue('master');
    await expect(form.locator('select').nth(1)).toHaveValue('6');
    await expect(form).toContainText('برای ریپوی عمومی لازم نیست');
    await expect(form.locator('[data-up-fork]')).toContainText('copilot-rast/nakama-farsi');
    await expect(token).toHaveAttribute('type', 'password');
    await shot(page, info, 'updates-settings');

    await repo.fill('https://example.com/a/b');
    await form.locator('[data-settings-save]').click();
    await expect(form).toContainText('نشانی ریپو درست نیست');
    await repo.fill('https://github.com/copilot-rast/nakama-farsi');
    await branch.fill('a..b');
    await form.locator('[data-settings-save]').click();
    await expect(form).toContainText('نام شاخه درست نیست');
    await branch.fill('main');
    await token.fill('short');
    await form.locator('[data-settings-save]').click();
    await expect(form).toContainText('این توکن شکل درستی ندارد');
    await token.fill(TOKEN);
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    await settle(page);
    await expect(repo).toHaveValue('copilot-rast/nakama-farsi');
    await expect(token).toHaveValue('********');
    await expect(form.locator('.eco-up-branch')).toHaveText(['main', 'stage', 'dev']);
    await form.locator('.eco-up-branch', { hasText: 'stage' }).click();
    await expect(branch).toHaveValue('stage');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');

    const saved = await base.get('/updates/settings');
    expect(saved).toMatchObject({ repo: 'copilot-rast/nakama-farsi', branch: 'stage', token: '********', hasToken: true });
    expect(JSON.stringify(saved)).not.toContain(TOKEN);
    const sealed = sql(`SELECT token_sealed FROM base_updates_settings`);
    expect(sealed.length).toBeGreaterThan(20);
    expect(sealed).not.toContain('e2eAAAA');
    const audit = sql(`SELECT string_agg(metadata::text, ' ') FROM console_audit_log WHERE message LIKE 'base update updates/settings%' AND create_time > now() - interval '5 minutes'`);
    expect(audit).toContain('********');
    expect(audit).not.toContain(TOKEN);
    expect(sql(`SELECT count(*) FROM console_audit_log WHERE metadata::text LIKE '%${TOKEN}%'`)).toBe('0');

    await settled((x) => x.state.status === 'base_mismatch' && !x.state.queued);
    await open(page, '#/system/updates');
    await expect(root(page).locator('[data-up-notice="base_mismatch"]')).toContainText('heroiclabs/nakama');
    await expect(root(page).locator('[data-up-guide]')).toContainText('git fetch upstream stage');
    await expect(root(page).locator('[data-up-guide]')).toContainText('فورکی از copilot-rast/nakama-farsi');
    await expect(root(page).locator('[data-up-commit]')).toHaveCount(0);
    expect((await base.get('/updates/summary')).newCommits).toBe(0);

    const bad = await consoleApi('PUT', '/v2/console/base/updates/settings', { repo: 'heroiclabs/nakama', branch: 'master', intervalHours: 500 });
    expect(bad.status).toBe(400);
    expect(bad.data.error).toBe('INVALID_INTERVAL');
    await base.put('/updates/settings', { enabled: true, provider: 'github', repo: 'heroiclabs/nakama', branch: 'master', intervalHours: 6, token: '' });
    expect((await base.get('/updates/settings')).hasToken).toBe(false);
    sql(`DELETE FROM base_updates_settings`);
    await fresh();
  });
});
