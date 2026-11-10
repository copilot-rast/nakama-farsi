import type { Page } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { open, openSettings, root, rows, search, still } from '../../lib/console';
import { base, consoleApi, eco } from '../../lib/api';
import { restoreConfig } from '../../lib/db';

const LATIN_UI = /Confirm Deletion|Are you sure|Confirm interrupt|Confirm Action/;

async function native(page: Page, hash: string, ready = '#main > main') {
  await page.goto('/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.locator(ready).first().waitFor({ state: 'attached' });
  await page.waitForFunction(() => !document.querySelector('#main .animate-pulse'));
}

async function charts(page: Page) {
  return page.evaluate(() => {
    const out: any[] = [];
    const seen = new Set();
    const walkV = (v: any) => {
      if (!v) return;
      if (v.component) walkI(v.component);
      if (Array.isArray(v.children)) v.children.forEach((c: any) => c && typeof c === 'object' && walkV(c));
      if (v.suspense && v.suspense.activeBranch) walkV(v.suspense.activeBranch);
    };
    const walkI = (i: any) => {
      if (seen.has(i)) return;
      seen.add(i);
      if (i.type && i.type.name === 'echarts' && i.setupState.chart) {
        const o = i.setupState.chart.getOption();
        out.push({ inverse: o.xAxis[0].inverse, y: o.yAxis[0].position, title: o.title && o.title[0] && o.title[0].left });
      }
      walkV(i.subTree);
    };
    walkV((document.getElementById('app') as any)._vnode);
    return out;
  });
}

test.describe('audit fixes', () => {
  test('English console on a Persian browser shows Latin digits, Gregorian dates and real KPI numbers', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, locale: 'en' });
    await native(page, '#/', 'number-flow-vue');
    await expect.poll(() => page.evaluate(() => {
      const flows = [...document.querySelectorAll('#main number-flow-vue')];
      const bad = flows.some((n) => [...(n.shadowRoot?.querySelectorAll('[part~="digit"]') || [])].some((d) => isNaN(parseInt((d as HTMLElement).style.getPropertyValue('--current'), 10))));
      return flows.length > 0 && !bad;
    })).toBe(true);
    const text = await page.locator('#main').innerText();
    expect(text).not.toMatch(/[۰-۹]/);
    expect(await page.evaluate(() => new Date(Date.UTC(2024, 2, 20, 12)).toLocaleDateString(undefined, { timeZone: 'UTC' }))).toBe('3/20/2024');
    await page.locator('.sp-status').click();
    await expect(page.locator('.sp-stats dd').first()).toHaveText(/^\d+$/);
    await expect(page.locator('.sp-stats')).not.toContainText('–');
    const box = await page.locator('.sp-stats').boundingBox();
    expect(box!.x + box!.width).toBeLessThanOrEqual(1440);
    expect(box!.x).toBeGreaterThanOrEqual(0);
  });

  test('native delete confirmation is Persian and the modal dims the page', async ({ page }) => {
    await native(page, '#/players', '#main tbody tr');
    await page.locator('#main tbody tr').nth(1).locator('button').last().click();
    await page.locator('.v-popper__popper--shown').getByText('حذف').last().click();
    const form = page.locator('.x-modal .x-form').filter({ hasText: 'تأیید حذف' });
    await expect(form).toBeVisible();
    await expect(form).toContainText('مطمئن هستید؟');
    expect(await form.innerText()).not.toMatch(LATIN_UI);
    const scrim = await page.evaluate(() => getComputedStyle(document.querySelector('.x-modal > div.fixed.inset-0')!).backgroundColor);
    const alpha = Number((/rgba?\([^)]*?,\s*([\d.]+)\)$/.exec(scrim) || [0, 1])[1]);
    expect(alpha).toBeGreaterThanOrEqual(0.25);
    await page.keyboard.press('Escape');
    await expect(form).toHaveCount(0);
  });

  test('Monaco follows dark mode and joins Persian letters', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, mode: 'dark' });
    await native(page, '#/storage', '#main tbody tr');
    await page.locator('#main tbody tr').first().click();
    const bg = page.locator('.monaco-editor .monaco-editor-background').first();
    await bg.waitFor();
    const color = await bg.evaluate((el) => getComputedStyle(el).backgroundColor);
    const [r, g, b] = (color.match(/\d+/g) || []).map(Number);
    expect(0.2126 * r + 0.7152 * g + 0.0722 * b).toBeLessThan(60);
    const font = await page.locator('.monaco-editor .view-lines').first().evaluate((el) => getComputedStyle(el).fontFamily);
    expect(font).toMatch(/^"?Vazirmatn/);
    await page.keyboard.press('Escape');
  });

  test('left-to-right fields keep their direction and show https:// in order', async ({ page }) => {
    await openSettings(page, 'brand');
    const url = root(page).locator('input[placeholder="https://"]').first();
    await expect(url).toBeVisible();
    expect(await url.evaluate((el) => [el.getAttribute('dir'), getComputedStyle(el).direction, getComputedStyle(el, '::placeholder').direction])).toEqual(['ltr', 'ltr', 'ltr']);
    const hex = root(page).locator('input.font-mono[dir="ltr"]').first();
    expect(await hex.evaluate((el) => getComputedStyle(el).direction)).toBe('ltr');
  });

  test('runtime module paths keep the slash in front and dates are localized', async ({ page }) => {
    await native(page, '#/runtime', '#main table tbody tr');
    const cell = page.locator('#main td [data-l10n-token], #main td[data-l10n-token]').filter({ hasText: /\.so$/ }).first();
    await expect(cell).toBeVisible();
    expect(await cell.evaluate((el) => getComputedStyle(el).unicodeBidi)).toBe('plaintext');
    const text = await page.locator('#main').innerText();
    expect(text).not.toMatch(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/);
  });

  test('dashboard charts run right to left in Persian and left to right in English', async ({ page, spawn }) => {
    await native(page, '#/', '[_echarts_instance_]');
    await expect.poll(async () => (await charts(page)).length).toBeGreaterThan(0);
    for (const c of await charts(page)) expect(c).toMatchObject({ inverse: true, y: 'right', title: 'right' });
    const en = await spawn({ signedIn: true, locale: 'en' });
    await native(en, '#/', '[_echarts_instance_]');
    await expect.poll(async () => (await charts(en)).length).toBeGreaterThan(0);
    for (const c of await charts(en)) expect(c).toMatchObject({ inverse: false, y: 'left' });
  });

  test('economy and publishing tab titles are translated and name the page', async ({ page, spawn }) => {
    await open(page, '#/economy/products');
    await expect(page).toHaveTitle(/^اقتصاد - فروشگاه \|/);
    await open(page, '#/publication/news');
    await expect(page).toHaveTitle(/^انتشار - اخبار \|/);
    const en = await spawn({ signedIn: true, locale: 'en' });
    await open(en, '#/economy/products');
    await expect(en).toHaveTitle(/^Economy - Shop \|/);
  });

  test('a coin price shows the currency name, not its key', async ({ page }) => {
    const products = (await eco.get('/products?limit=100')).items as any[];
    const currencies = (await eco.get('/currencies?limit=100')).items as any[];
    const p = products.find((x) => x.coinPrice && currencies.some((c) => c.key === x.coinPrice.currencyKey && c.name && c.name.fa));
    test.skip(!p, 'no product with a coin price');
    const cur = currencies.find((c) => c.key === p.coinPrice.currencyKey);
    await open(page, '#/economy/products');
    await search(page, p.key);
    const row = rows(page).filter({ hasText: p.key }).first();
    await expect(row).toContainText(cur.name.fa);
    expect(await row.locator('td').nth(1).innerText()).not.toContain(cur.key);
  });

  test('the chat page does not fire a failing request and the Satori tab is hidden when Satori is off', async ({ page }) => {
    const bad: string[] = [];
    page.on('response', (r) => { if (r.status() >= 400 && /\/v2\/console\/(channel|satori)/.test(r.url())) bad.push(r.status() + ' ' + r.url()); });
    await native(page, '#/chat');
    await page.waitForTimeout(800);
    const ext = await consoleApi('GET', '/v2/console/extensions');
    if (!ext.data.satori) {
      await native(page, '#/players', '#main tbody tr');
      await page.locator('#main tbody tr').nth(1).click();
      await expect(page).toHaveURL(/#\/players\/[^/]+/);
      await expect(page.locator('#main a[href*="/satori"]')).toHaveCount(0);
      const id = decodeURIComponent(new URL(page.url()).hash.split('/')[2]);
      await page.evaluate((h) => { location.hash = h; }, '#/players/' + id + '/satori/messages');
      await expect(page).not.toHaveURL(/satori/);
      await expect(page).toHaveURL(new RegExp('#/players/' + id));
    }
    expect(bad).toEqual([]);
  });

  test('an unknown page says it was not found', async ({ page }) => {
    await native(page, '#/players');
    await page.goto('/#/does-not-exist-e2e');
    await expect(page.locator('.eco-toasts')).toContainText('پیدا نشد');
    await expect(page).toHaveURL(/#\/$/);
  });

  test('page search ranks page names before section names', async ({ page }) => {
    await native(page, '#/');
    await page.keyboard.press('Control+k');
    await page.locator('.sp-pal-input').fill('تن');
    const labels = await page.locator('.sp-pal-row .sp-pal-label').allInnerTexts();
    expect(labels.length).toBeGreaterThan(0);
    for (const l of labels) expect(l).toContain('تن');
    await page.keyboard.press('Escape');
  });

  test('module rows in the audit log are not labelled Hiro', async ({ page }) => {
    await native(page, '#/settings/audit-log', '#main table tbody tr');
    await page.waitForTimeout(500);
    const mod = page.locator('#main tbody tr').filter({ hasText: /^(?:[\s\S]*)(پایه|اقتصاد|انتشار) · / });
    if (await mod.count()) {
      for (const t of await mod.allInnerTexts()) expect(t).not.toMatch(/Hiro/);
    }
    expect(await page.locator('#main tbody').innerText()).not.toMatch(/^(base|economy|publication) (create|update|delete|invoke) /m);
  });

  test('the dashboard offers to continue an unfinished setup', async ({ page }) => {
    try {
      await base.put('/setup', { status: 'pending' });
      await base.put('/setup', { status: 'skipped', steps: { brand: 'done', languages: 'done' } });
      await native(page, '#/', '.grid.grid-cols-1.gap-6');
      const card = page.locator('[data-setup-nudge]');
      await expect(card).toBeVisible();
      await expect(card).toContainText('برای بعد');
      await expect(card).toContainText('۲ از ۶');
      await card.locator('[data-setup-nudge-go]').click();
      await expect(page.locator('.setup-overlay')).toBeVisible();
      await page.locator('[data-setup-later]').click();
      await expect(page.locator('.setup-overlay')).toHaveCount(0);
    } finally {
      restoreConfig();
    }
  });

  test('setup page keeps the user and date apart', async ({ page }) => {
    await openSettings(page, 'setup');
    const dd = root(page).locator('.eco-dl dd').first();
    if (await dd.count()) {
      expect(await dd.locator('bdi').count()).toBeGreaterThanOrEqual(1);
    }
    const desc = root(page).locator('[data-setup-row] .text-xs').first();
    expect(await desc.evaluate((el) => getComputedStyle(el).whiteSpace)).not.toBe('nowrap');
  });
});

test.describe('audit fixes on a phone', () => {
  test.use({ pageOpts: { signedIn: true, mobile: true } } as any);

  test('config export button and storage IDs stay inside the screen', async ({ page }) => {
    await native(page, '#/settings/config', '#main .rounded-md.shadow');
    const btn = page.locator('#main .translate-x-1.shrink-0 .x-button').first();
    await expect(btn).toBeVisible();
    const b = await btn.boundingBox();
    expect(b!.x).toBeGreaterThanOrEqual(0);
    expect(b!.x + b!.width).toBeLessThanOrEqual(390);
    await native(page, '#/storage', '#main tbody tr');
    const over = await page.evaluate(() => [...document.querySelectorAll('#main table[data-sp-cards] td[data-sp-label]')].filter((td) => td.scrollWidth > td.clientWidth + 1).length);
    expect(over).toBe(0);
  });

  test('economy rows are framed cards like native ones', async ({ page }) => {
    await open(page, '#/economy/products');
    const tr = root(page).locator('table.eco-table tbody tr').first();
    await expect(tr).toBeVisible();
    const style = await tr.evaluate((el) => ({ border: getComputedStyle(el).borderTopWidth, radius: getComputedStyle(el).borderTopLeftRadius }));
    expect(style.border).toBe('1px');
    expect(parseFloat(style.radius)).toBeGreaterThan(4);
    await still(page);
  });
});
