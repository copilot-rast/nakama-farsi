import fs from 'node:fs';
import path from 'node:path';
import type { Page } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { measureContrast } from '../../lib/contrast';
import { noHorizontalScroll, still } from '../../lib/console';
import { ROOT } from '../../lib/env';
import { restoreConfig } from '../../lib/db';

const SELECTORS = JSON.parse(fs.readFileSync(path.join(ROOT, '..', 'console', 'ui', 'plane', 'selectors.json'), 'utf8'));
const ZERO = '00000000-0000-0000-0000-000000000000';

type Screen = { name: string; hash: string; ready?: string; act?: (page: Page) => Promise<void> };

async function firstId(page: Page, list: string, pattern: RegExp) {
  await page.goto('/' + list);
  await page.locator('#main tbody tr').first().waitFor();
  await page.locator('#main tbody tr').first().click();
  await expect(page).toHaveURL(pattern);
  return decodeURIComponent(new URL(page.url()).hash.split('/')[2]);
}

async function show(page: Page, hash: string, ready = '#main > main') {
  await page.goto('/' + hash);
  await page.locator('.pl-rail').waitFor();
  await page.locator(ready).first().waitFor();
  await page.waitForFunction(() => !document.querySelector('#main .animate-pulse, [data-economy] .animate-pulse'));
  await page.waitForTimeout(400);
  await still(page);
}

async function check(page: Page, testInfo: any, name: string, mode: string) {
  await noHorizontalScroll(page);
  const c = await measureContrast(page);
  const shot = await page.screenshot({ type: 'jpeg', quality: 60 });
  await testInfo.attach(`${name}-${mode}.jpg`, { body: shot, contentType: 'image/jpeg' });
  expect(c.fails, `${name} ${mode}: text under 4.5:1 (min ${c.min.toFixed(2)} of ${c.count} runs)`).toEqual([]);
  expect(c.count).toBeGreaterThan(3);
  const transformed = await page.evaluate(() => {
    const bad: string[] = [];
    const main = document.getElementById('main');
    if (main && getComputedStyle(main).transform !== 'none') bad.push('#main');
    return bad;
  });
  expect(transformed).toEqual([]);
}

const NATIVE: Record<string, Screen[]> = {
  lists: [
    { name: 'players', hash: '#/players' },
    { name: 'groups', hash: '#/groups' },
    { name: 'storage', hash: '#/storage' },
    { name: 'storage-indexes', hash: '#/storage-indexes' },
    { name: 'matches', hash: '#/matches' },
    { name: 'leaderboards', hash: '#/leaderboards' },
    { name: 'notifications', hash: '#/notifications' },
    { name: 'purchases', hash: '#/payments' },
    { name: 'subscriptions', hash: '#/payments/subscriptions' },
    { name: 'chat', hash: '#/chat' },
    { name: 'runtime', hash: '#/runtime' },
    { name: 'audit-log', hash: '#/settings/audit-log' },
    { name: 'config', hash: '#/settings/config' },
  ],
  details: [
    { name: 'player-account', hash: `#/players/${ZERO}` },
    { name: 'player-wallet', hash: `#/players/${ZERO}/wallet` },
    { name: 'player-storage', hash: `#/players/${ZERO}/storage` },
    { name: 'player-friends', hash: `#/players/${ZERO}/friends` },
    { name: 'player-groups', hash: `#/players/${ZERO}/groups` },
    { name: 'player-payments', hash: `#/players/${ZERO}/payments` },
    { name: 'player-notifications', hash: `#/players/${ZERO}/notifications` },
    { name: 'player-hiro', hash: `#/players/${ZERO}/hiro` },
    { name: 'player-hiro-inventory', hash: `#/players/${ZERO}/hiro/inventory` },
    { name: 'player-satori', hash: `#/players/${ZERO}/satori/messages` },
  ],
  tools: [
    { name: 'dashboard', hash: '#/' },
    { name: 'api-explorer', hash: '#/api-explorer' },
    { name: 'settings-users', hash: '#/settings/users' },
    { name: 'settings-general', hash: '#/settings/general' },
    { name: 'data-management', hash: '#/settings/data-management' },
  ],
};

test.describe('plane visual: native pages', () => {
  for (const [group, screens] of Object.entries(NATIVE)) {
    for (const mode of ['light', 'dark'] as const) {
      test(`${group} in fa ${mode}: no sideways scroll, text at 4.5:1`, async ({ spawn }, testInfo) => {
        const page = await spawn({ signedIn: true, mode });
        for (const s of screens) {
          await show(page, s.hash);
          await check(page, testInfo, s.name, mode);
        }
      });
    }
  }

  test('group detail and members in fa light and en dark', async ({ spawn }, testInfo) => {
    const page = await spawn({ signedIn: true });
    const id = await firstId(page, '#/groups', /#\/groups\/[0-9a-f-]+/);
    for (const hash of [`#/groups/${id}`, `#/groups/${id}/members`]) {
      await show(page, hash);
      await check(page, testInfo, hash.split('/').slice(-1)[0], 'light');
    }
    const en = await spawn({ signedIn: true, locale: 'en', mode: 'dark' });
    await show(en, `#/groups/${id}/members`);
    await expect(en.locator('html')).toHaveAttribute('dir', 'ltr');
    await check(en, testInfo, 'members-en', 'dark');
  });

  for (const mode of ['light', 'dark'] as const) {
    test(`overlays in fa ${mode}: row menu, editor windows and a confirm dialog are readable`, async ({ spawn }, testInfo) => {
      const page = await spawn({ signedIn: true, mode });
      await show(page, '#/players');
      await page.locator('#main tbody tr').nth(1).locator('button').last().click();
      await expect(page.locator('.v-popper__popper--shown')).toBeVisible();
      await still(page);
      await check(page, testInfo, 'row-menu', mode);
      await page.keyboard.press('Escape');
      await show(page, '#/settings/users');
      await page.locator('#main button', { hasText: 'ایجاد کاربر' }).click();
      const form = page.locator('div.fixed.inset-0.flex.justify-end > .x-form');
      await expect(form).toBeVisible();
      await still(page);
      await check(page, testInfo, 'user-add', mode);
      await show(page, '#/groups');
      await page.locator('#main tbody tr').first().locator('button').last().click();
      await page.locator('.v-popper__popper--shown').getByText('حذف').last().click();
      const modal = page.locator('.x-modal form.x-form');
      await expect(modal).toBeVisible();
      await still(page);
      expect(await modal.evaluate((e) => getComputedStyle(e).borderTopLeftRadius)).toBe('8px');
      await check(page, testInfo, 'confirm', mode);
    });
  }

  test('every manifest selector still matches on its routes', async ({ page }) => {
    for (const entry of SELECTORS.selectors) {
      for (const hash of entry.routes) {
        await show(page, hash, entry.id === 'auth' ? 'body' : '#main > main');
        if (entry.id === 'auth') continue;
        const n = await page.locator(entry.selector).count();
        expect(n, `${entry.id} on ${hash}: ${entry.selector}`).toBeGreaterThanOrEqual(entry.min);
      }
    }
  });

  test('dashboard KPI strip and charts: one window, four SVG charts themed by tokens', async ({ page }) => {
    await show(page, '#/', '[_echarts_instance_] svg');
    const strip = page.locator('#main .grid.xl\\:grid-cols-5');
    await expect(strip.locator(':scope > .rounded-md.shadow.bg-white')).toHaveCount(5);
    expect(await strip.evaluate((e) => getComputedStyle(e).borderRadius)).toBe('8px');
    await expect(page.locator('#main [_echarts_instance_] svg')).toHaveCount(4);
    const fill = await page.locator('#main [_echarts_instance_] text').first().evaluate((e) => getComputedStyle(e).fill);
    const ink3 = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--ink-3').trim());
    expect(ink3).toBe('oklch(0.4176 0.0072 239.98)');
    expect(fill.replace(/\s+/g, ' ')).toBe(ink3);
  });

  test('storage object editor: docked panel, Monaco takes typing where it was clicked', async ({ page }) => {
    await show(page, '#/storage');
    const row = page.locator('#main tbody tr').filter({ hasText: 'economy_items' }).first();
    await row.click();
    const form = page.locator('div.fixed.inset-0.flex.justify-end > .x-form');
    await expect(form).toBeVisible();
    await still(page);
    const r = await form.evaluate((e) => ({ radius: getComputedStyle(e).borderTopLeftRadius, top: e.getBoundingClientRect().top }));
    expect(r.radius).toBe('0px');
    expect(r.top).toBeGreaterThanOrEqual(0);
    const editor = form.locator('.monaco-editor').first();
    await expect(editor).toBeVisible();
    const box = await editor.locator('.view-lines').boundingBox();
    await page.mouse.click(box!.x + 4, box!.y + 4);
    await page.keyboard.press(process.platform === 'darwin' ? 'Meta+End' : 'Control+End');
    const before = await editor.locator('.view-lines').innerText();
    await page.keyboard.type(' ');
    await expect.poll(() => editor.locator('.view-lines').innerText()).not.toBe(before);
    await page.keyboard.press('Escape');
  });
});

const OURS: Record<string, string[]> = {
  economy: ['#/economy/overview', '#/economy/currencies', '#/economy/metas', '#/economy/prizes', '#/economy/products', '#/economy/items', '#/economy/achievements', '#/economy/chance-games', '#/economy/prize-groups', '#/economy/games', '#/economy/gifts', '#/economy/event-types', '#/economy/event-rules', '#/economy/meta-rules', '#/economy/level-tables', '#/economy/streaks', '#/economy/tasks', '#/economy/task-groups', '#/economy/event-log', '#/economy/payments', '#/economy/products/new', '#/economy/level-tables/new', '#/economy/streaks/new', '#/economy/tasks/new', '#/economy/task-groups/new', '#/economy/items/new', '#/economy/achievements/new', '#/economy/chance-games/new', '#/economy/prize-groups/new', '#/economy/games/new'],
  publication: ['#/publication/faqs', '#/publication/news', '#/publication/banners', '#/publication/banner-lists', '#/publication/banners/new', '#/publication/sponsors', '#/publication/sponsors/new', '#/publication/auto-rules', '#/publication/auto-rules/new', '#/publication/notify-categories', '#/publication/comments', '#/publication/notifications', '#/publication/notifications/new'],
  moderation: ['#/moderation/players', '#/moderation/reports', '#/moderation/restrictions', '#/moderation/chat-words', '#/moderation/chat-words/new'],
  system: ['#/system/updates'],
  settings: ['#/project-settings/brand', '#/project-settings/setup', '#/project-settings/languages', '#/project-settings/countries', '#/project-settings/accounts', '#/project-settings/otp', '#/project-settings/app-versions', '#/project-settings/categories', '#/project-settings/tags', '#/project-settings/report-titles', '#/project-settings/events', '#/project-settings/payment-methods', '#/project-settings/sms', '#/project-settings/email', '#/project-settings/updates-source'],
};

test.describe('plane visual: our pages', () => {
  for (const [group, hashes] of Object.entries(OURS)) {
    for (const mode of ['light', 'dark'] as const) {
      test(`${group} in fa ${mode}: glass windows and readable text`, async ({ spawn }, testInfo) => {
        const page = await spawn({ signedIn: true, mode });
        for (const hash of hashes) {
          await show(page, hash, hash.includes('project-settings') ? '#project-settings-body' : '#economy-root');
          await check(page, testInfo, hash.replace(/[#/]+/g, '-').slice(1), mode);
        }
      });
    }
  }

  test('list pages put their header inside one window; overview is three KPI strips', async ({ page }) => {
    await show(page, '#/economy/products', '#economy-root table');
    const win = page.locator('#economy-root div.max-w-\\[1600px\\][data-pl-list]');
    await expect(win).toHaveCount(1);
    await expect(win.locator('h1')).toHaveText('فروشگاه');
    expect(await win.evaluate((e) => getComputedStyle(e).borderTopLeftRadius)).toBe('8px');
    await show(page, '#/economy/overview', '#economy-root .eco-stat-grid');
    const strips = page.locator('#economy-root .eco-stat-grid');
    await expect(strips).toHaveCount(3);
    expect(await strips.first().evaluate((e) => getComputedStyle(e).borderTopLeftRadius)).toBe('8px');
  });

  test('product editor: language tabs as a full-width segment, payment methods as chips, docked panel', async ({ page }) => {
    await show(page, '#/economy/products/new', '[data-economy="panel"] form');
    const form = page.locator('[data-economy="panel"] .x-form');
    await still(page);
    const w = await form.evaluate((e) => ({ width: e.getBoundingClientRect().width, radius: getComputedStyle(e).borderTopLeftRadius }));
    expect(w.width).toBeLessThanOrEqual(682);
    expect(w.radius).toBe('0px');
    const tabs = form.locator('[data-translated="name"] .eco-tr-tabs');
    const tb = await tabs.boundingBox();
    const inner = await form.locator('[data-translated="name"]').boundingBox();
    expect(Math.abs(tb!.width - inner!.width)).toBeLessThan(2);
    await tabs.locator('[data-lang="en"]').click();
    await expect(tabs.locator('[data-lang="en"]')).toHaveAttribute('aria-selected', 'true');
    expect(await tabs.locator('[data-lang="en"]').evaluate((e) => getComputedStyle(e).boxShadow)).not.toBe('none');
    const chip = form.locator('.eco-choices .x-checkbox').nth(1);
    const before = await chip.getAttribute('aria-checked');
    await chip.click();
    await expect(chip).toHaveAttribute('aria-checked', before === 'true' ? 'false' : 'true');
    expect(await chip.evaluate((e) => getComputedStyle(e).borderTopLeftRadius)).toBe('6px');
  });

  test('setup wizard sits on the canvas with one centered window and a segmented step bar', async ({ page }, testInfo) => {
    await show(page, '#/project-settings/setup', '#project-settings-body');
    try {
      await page.locator('#project-settings-body button', { hasText: 'باز کردن راهنما' }).click();
      const inner = page.locator('.setup-overlay .setup-inner');
      await expect(inner).toBeVisible();
      await still(page);
      const r = await inner.evaluate((e) => ({ radius: getComputedStyle(e).borderTopLeftRadius, w: e.getBoundingClientRect().width }));
      expect(r.radius).toBe('12px');
      expect(r.w).toBeLessThanOrEqual(762);
      await expect(page.locator('.setup-step--on')).toHaveCount(1);
      await check(page, testInfo, 'wizard', 'light');
    } finally {
      restoreConfig();
    }
  });
});
