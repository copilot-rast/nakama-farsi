import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, must, pub, uid } from '../../lib/api';
import { ADMIN, field, loginAs, noHorizontalScroll, open, openSettings, panel, psWindow, root, settle, settlePanel, toasts } from '../../lib/console';
import { CONSOLE_URL } from '../../lib/env';
import type { Page } from '@playwright/test';

const k = uid();
const created = { categories: [] as string[], tags: [] as string[] };

test.afterAll(async () => {
  for (const id of created.categories) await base.del('/categories/' + id);
  for (const id of created.tags) await base.del('/tags/' + id);
});

const SECTIONS: [string, string, string][] = [
  ['brand', 'برند', '.brand-editor'],
  ['setup', 'راه‌اندازی', '[data-setup-row]'],
  ['languages', 'زبان‌ها', 'table.eco-table'],
  ['countries', 'کشورها', 'table.eco-table'],
  ['accounts', 'قوانین حساب', '[data-settings-form="accounts"] form'],
  ['otp', 'ورود با پیامک', '[data-settings-form="otp"] form'],
  ['app-versions', 'نسخه‌های برنامه', 'table.eco-table'],
  ['categories', 'دسته‌بندی‌ها', 'table.eco-table'],
  ['tags', 'تگ‌ها', 'table.eco-table'],
  ['report-titles', 'عنوان‌های گزارش', 'table.eco-table'],
  ['payment-methods', 'درگاه‌های پرداخت', '[data-method]'],
  ['sms', 'پیامک', '[data-channel="sms"]'],
  ['email', 'ایمیل', '[data-channel="email"]'],
];

function tab(page: Page, id: string) {
  return psWindow(page).locator(`[role="tab"][data-ps-section="${id}"]`);
}

async function findByKey(resource: 'categories' | 'tags', key: string) {
  const r = await base.get(`/${resource}?q=${encodeURIComponent(key)}&limit=20`);
  return r.items.find((x: any) => x.key === key);
}

async function byName(resource: 'categories' | 'tags', name: string) {
  const r = await base.get(`/${resource}?q=${encodeURIComponent(name)}&limit=20`);
  return r.items.filter((x: any) => x.name?.fa === name);
}

async function quickCreate(page: Page, label: RegExp, resource: 'categories' | 'tags', name: string) {
  const f = field(panel(page), label);
  await f.locator(`[data-quick-create="${resource}"]`).click();
  const pop = f.locator('.eco-qc-pop');
  await expect(pop).toBeVisible();
  await expect(pop.locator('input')).toBeFocused();
  await pop.locator('input').fill(name);
  await pop.locator('[data-qc-create]').click();
  await expect(pop).toBeHidden();
  await expect(toasts(page).filter({ hasText: 'ساخته و انتخاب شد' }).last()).toBeVisible();
  const items = await byName(resource, name);
  expect(items.length).toBeGreaterThan(0);
  for (const it of items) if (!created[resource].includes(it.id)) created[resource].push(it.id);
  return items[items.length - 1];
}

test.describe('project settings window', () => {
  test('opens from the top bar gear and from the sidebar, Esc and close return to the page underneath', async ({ page }) => {
    await open(page, '#/economy/products');
    const gear = page.locator('[data-ps-entry="topbar"]');
    await expect(gear).toBeVisible();
    await gear.click();
    await expect(psWindow(page)).toBeVisible();
    await expect(page).toHaveURL(/#\/project-settings\/brand$/);
    await expect(psWindow(page).locator('#ps-title')).toHaveText('تنظیمات پروژه');
    await expect(tab(page, 'brand')).toBeFocused();
    await expect(page.locator('.x-drawer [data-ps-entry="sidebar"]').first()).toHaveAttribute('aria-current', 'page');
    await expect(page.locator('.sp-rail button[data-sp-group="settings"]')).toHaveAttribute('aria-current', 'true');
    await page.keyboard.press('Escape');
    await expect(psWindow(page)).toHaveCount(0);
    await expect(page).toHaveURL(/#\/economy\/products$/);
    await expect(root(page).locator('h1')).toHaveText('فروشگاه');

    await page.locator('.sp-tabs').waitFor();
    await page.locator('.sp-rail button[data-sp-group="publication"]').click();
    await page.locator('.sp-flyout.open a[href^="#/publication/news"]').click();
    await expect(root(page).locator('h1')).toHaveText('اخبار');
    await page.locator('.sp-rail button[data-sp-group="settings"]').click();
    await page.locator('.sp-flyout.open [data-ps-entry="rail"]').click();
    await expect(psWindow(page)).toBeVisible();
    await tab(page, 'sms').click();
    await expect(page).toHaveURL(/#\/project-settings\/sms$/);
    await psWindow(page).locator('[data-ps-close]').click();
    await expect(psWindow(page)).toHaveCount(0);
    await expect(page).toHaveURL(/#\/publication\/news$/);
    await expect(root(page).locator('h1')).toHaveText('اخبار');
    await expect(page.locator('html')).not.toHaveAttribute('data-ps-open', '');
  });

  test('every section renders its panel with the right heading', async ({ page }) => {
    await openSettings(page, 'brand');
    for (const [id, title, marker] of SECTIONS) {
      await tab(page, id).click();
      await expect(page).toHaveURL(new RegExp('#/project-settings/' + id + '$'));
      await expect(tab(page, id)).toHaveAttribute('aria-selected', 'true');
      await expect(psWindow(page).locator('.ps-list [role="tab"][aria-selected="true"]')).toHaveCount(1);
      await expect(psWindow(page).locator('#ps-section-title')).toHaveText(title);
      await expect(root(page).locator(marker).first()).toBeVisible();
      await settle(page);
    }
    expect(await page.title()).toContain('تنظیمات پروژه');
  });

  test('deep link survives a reload; section switches replace history so back closes the window', async ({ page }) => {
    await open(page, '#/economy/currencies');
    await page.locator('[data-ps-entry="topbar"]').click();
    await expect(psWindow(page)).toBeVisible();
    await tab(page, 'tags').click();
    await tab(page, 'sms').click();
    await expect(page).toHaveURL(/#\/project-settings\/sms$/);
    await page.reload();
    await expect(psWindow(page)).toBeVisible();
    await expect(tab(page, 'sms')).toHaveAttribute('aria-selected', 'true');
    await settle(page);
    await expect(root(page).locator('[data-channel="sms"]')).toBeVisible();
    await page.goBack();
    await expect(page).toHaveURL(/#\/economy\/currencies$/);
    await expect(psWindow(page)).toHaveCount(0);
    await expect(root(page).locator('h1')).toHaveText('ارزها');

    await open(page, '#/project-settings/categories?keyType=prize');
    await expect(root(page).locator('.eco-filters select').first()).toHaveValue('prize');
    await open(page, '#/project-settings/nope');
    await expect(page).toHaveURL(/#\/project-settings\/brand$/);
  });

  test('keyboard: arrows move between sections, Tab stays inside, Esc inside an editor closes only the editor', async ({ page }) => {
    await openSettings(page, 'brand');
    await tab(page, 'brand').focus();
    await page.keyboard.press('ArrowDown');
    await expect(tab(page, 'setup')).toBeFocused();
    await expect(page).toHaveURL(/#\/project-settings\/setup$/);
    await page.keyboard.press('End');
    await expect(tab(page, 'email')).toBeFocused();
    await expect(page).toHaveURL(/#\/project-settings\/email$/);
    await page.keyboard.press('Home');
    await expect(page).toHaveURL(/#\/project-settings\/brand$/);
    await page.keyboard.press('ArrowUp');
    await expect(page).toHaveURL(/#\/project-settings\/email$/);
    await expect(root(page).locator('[data-channel="email"]')).toBeVisible();
    await settle(page);
    for (let i = 0; i < 25; i++) {
      await page.keyboard.press('Tab');
      expect(await page.evaluate(() => !!document.activeElement?.closest('.ps-window'))).toBe(true);
    }
    await tab(page, 'categories').click();
    await settle(page);
    await root(page).getByRole('button', { name: 'دسته‌بندی جدید' }).first().click();
    await settlePanel(page);
    await expect(page).toHaveURL(/#\/project-settings\/categories\/new/);
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);
    await expect(psWindow(page)).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(psWindow(page)).toHaveCount(0);
  });

  test('a limited console user only sees the sections their ACL allows', async ({ spawn }) => {
    const pass = () => 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!';
    const eco = { user: 'e2eps' + k.slice(4, 12), pass: pass() };
    const none = { user: 'e2epn' + k.slice(4, 12), pass: pass() };
    async function addUser(who: { user: string; pass: string }, acl: Record<string, unknown>) {
      const add = await consoleApi('POST', '/v2/console/user', { username: who.user, email: who.user + '@example.com', acl });
      expect(add.status).toBeLessThan(300);
      const reset = await must(consoleApi('POST', '/v2/console/user/' + who.user + '/reset/password'));
      const code = reset.code || add.data?.token;
      const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code, password: who.pass }) });
      expect(pc.ok).toBe(true);
    }
    try {
      await addUser(eco, { ACCOUNT: { read: true }, HIRO_ECONOMY: { read: true, write: true } });
      await addUser(none, { ACCOUNT: { read: true } });

      const p = await spawn({ signedIn: false });
      await loginAs(p, eco);
      await p.waitForTimeout(800);
      await expect(p.locator('[data-ps-entry="topbar"]')).toHaveAttribute('href', '#/project-settings/categories');
      await p.locator('[data-ps-entry="topbar"]').click();
      await expect(psWindow(p)).toBeVisible();
      const ids = await psWindow(p).locator('.ps-list [role="tab"]').evaluateAll((els) => els.map((e) => e.getAttribute('data-ps-section')));
      expect(ids).toEqual(['categories', 'tags', 'events', 'referral-rules', 'payment-methods']);
      await settle(p);
      await expect(root(p).locator('table.eco-table').first()).toBeVisible();
      await p.goto(CONSOLE_URL + '/#/project-settings/sms');
      await expect(p).toHaveURL(/#\/project-settings\/categories$/);
      await p.goto(CONSOLE_URL + '/#/settings/brand');
      await expect(p).toHaveURL(/#\/project-settings\/categories$/);
      await expect(psWindow(p).locator('[data-ps-section="brand"]')).toHaveCount(0);

      const q = await spawn({ signedIn: false });
      await loginAs(q, none);
      await q.waitForTimeout(800);
      await expect(q.locator('[data-ps-entry]')).toHaveCount(0);
      await q.goto(CONSOLE_URL + '/#/project-settings/brand');
      await expect(psWindow(q)).toBeVisible();
      await expect(psWindow(q)).toContainText('به هیچ بخشی از تنظیمات پروژه دسترسی ندارید.');
      await expect(psWindow(q).locator('.ps-list [role="tab"]')).toHaveCount(0);
      await q.keyboard.press('Escape');
      await expect(psWindow(q)).toHaveCount(0);
    } finally {
      await consoleApi('DELETE', '/v2/console/user?username=' + eco.user);
      await consoleApi('DELETE', '/v2/console/user?username=' + none.user);
    }
  });

  test('the edit drawer and its shade sit above the top bar and sticky table heads', async ({ page }) => {
    await open(page, '#/economy/prizes');
    await settle(page);
    await page.locator('#economy-root tbody tr.x-table-row').first().click();
    await settlePanel(page);
    const z = await page.evaluate(() => {
      const p = document.querySelector('[data-economy="panel"]')!;
      return [...p.children].map((c) => getComputedStyle(c).zIndex);
    });
    expect(z).toEqual(['60', '60']);
    const hits = await page.evaluate(() => {
      const form = document.querySelector('[data-economy="panel"] .x-form')!.getBoundingClientRect();
      const shadeX = document.documentElement.dir === 'rtl' ? window.innerWidth - 20 : 20;
      const pts: [number, number][] = [[form.left + form.width / 2, form.top + 15], [form.left + form.width / 2, form.top + 200], [shadeX, 20], [shadeX, 200]];
      return pts.map(([x, y]) => !!document.elementFromPoint(x, y)?.closest('[data-economy="panel"]'));
    });
    expect(hits).toEqual([true, true, true, true]);
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);
  });
});

test.describe('quick-create next to category and tag pickers', () => {
  test('product form: category and tag created in place and selected; empty name and duplicate names handled', async ({ page }) => {
    await open(page, '#/economy/products/new');
    await settlePanel(page);
    const catField = field(panel(page), /^دسته‌بندی$/);
    await catField.locator('[data-quick-create="categories"]').click();
    const pop = catField.locator('.eco-qc-pop');
    await expect(pop).toContainText('برای «محصول»، به فارسی.');
    await pop.locator('[data-qc-create]').click();
    await expect(pop.locator('.eco-qc-err')).toHaveText('یک نام بنویسید.');
    await page.keyboard.press('Escape');
    await expect(pop).toBeHidden();
    await expect(panel(page).locator('form')).toBeVisible();

    const catName = k + ' دسته سریع';
    const cat = await quickCreate(page, /^دسته‌بندی$/, 'categories', catName);
    expect(cat.keyType).toBe('product');
    expect(cat.key).toMatch(/^[a-z0-9][a-z0-9_.-]*$/);
    expect(cat.name).toEqual({ fa: catName });
    await expect(catField.locator('button[aria-haspopup="listbox"]')).toContainText(catName);

    const tagName = k + ' تگ سریع';
    const tg = await quickCreate(page, /^تگ‌ها$/, 'tags', tagName);
    expect(tg.keyType).toBe('product');
    await expect(field(panel(page), /^تگ‌ها$/).locator('.eco-tag', { hasText: tagName })).toBeVisible();

    const again = await quickCreate(page, /^تگ‌ها$/, 'tags', tagName);
    const both = await byName('tags', tagName);
    expect(both.length).toBe(2);
    expect(new Set(both.map((x: any) => x.key)).size).toBe(2);
    expect(again.keyType).toBe('product');
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);
  });

  test('prize, FAQ and news forms create with their own key type; a saved prize keeps the new category', async ({ page }) => {
    await open(page, '#/economy/prizes/new');
    await settlePanel(page);
    const prizeCat = await quickCreate(page, /^دسته‌بندی$/, 'categories', k + ' جایزه سریع');
    expect(prizeCat.keyType).toBe('prize');
    const prizeTag = await quickCreate(page, /^تگ‌ها$/, 'tags', k + ' تگ جایزه سریع');
    expect(prizeTag.keyType).toBe('prize');
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);

    await open(page, '#/publication/faqs/new');
    await settlePanel(page);
    const faqCat = await quickCreate(page, /^دسته‌بندی$/, 'categories', k + ' پرسش سریع');
    expect(faqCat.keyType).toBe('faq');
    await expect(field(panel(page), /^دسته‌بندی$/).locator('button[aria-haspopup="listbox"]')).toContainText(k + ' پرسش سریع');
    await page.keyboard.press('Escape');

    await open(page, '#/publication/news/new');
    await settlePanel(page);
    const newsCat = await quickCreate(page, /^دسته‌بندی$/, 'categories', k + ' خبر سریع');
    expect(newsCat.keyType).toBe('news');
    const f = field(panel(page), /^دسته‌بندی$/);
    await f.locator('button[aria-haspopup="listbox"]').click();
    await f.locator('.eco-pop input').fill(k);
    await expect(f.locator('[role="option"]')).toHaveCount(1);
    await expect(f.locator('[role="option"]').first()).toContainText(k + ' خبر سریع');
    await page.keyboard.press('Escape');
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);
    expect((await pub.get('/news?q=' + k)).total).toBe(0);
  });

  test('pickers are filtered by key type: player tags never show in content forms; the tags list knows the player type', async ({ page }) => {
    const player = await base.post('/tags', { key: k + '_ut', keyType: 'user', name: { fa: 'تگ بازیکن ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    const general = await base.post('/tags', { key: k + '_gt2', keyType: 'general', name: { fa: 'تگ همه‌جا ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    created.tags.push(player.id, general.id);
    const types = await base.get('/tags/key-types');
    expect(types.items.map((x: any) => x.key)).toContain('user');

    await open(page, '#/economy/products/new');
    await settlePanel(page);
    const tags = field(panel(page), /^تگ‌ها$/);
    await tags.locator('button[aria-haspopup="listbox"]').click();
    const search = page.waitForResponse((r) => r.url().includes('/tags?') && r.url().includes('q=' + k));
    await tags.locator('.eco-pop input').fill(k);
    await search;
    await expect(tags.locator('[role="option"]', { hasText: 'تگ همه‌جا ' + k })).toHaveCount(1);
    await expect(tags.locator('[role="option"]', { hasText: 'تگ بازیکن ' + k })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await page.keyboard.press('Escape');

    await open(page, '#/project-settings/tags');
    await root(page).locator('.eco-filters select').first().selectOption('user');
    await settle(page);
    await expect(page).toHaveURL(/keyType=user/);
    const row = root(page).locator('tbody tr.x-table-row').filter({ hasText: 'تگ بازیکن ' + k });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText('بازیکن');
    await root(page).getByRole('button', { name: 'تگ جدید' }).first().click();
    await settlePanel(page);
    await expect(field(panel(page), /^نوع کلید$/).locator('option[value="user"]')).toHaveText('بازیکن');
    await page.keyboard.press('Escape');
    expect(await findByKey('tags', k + '_ut')).toBeTruthy();
  });

  test('quick-create works in English and dark mode', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, locale: 'en', mode: 'dark' });
    await open(page, '#/economy/products/new');
    await settlePanel(page);
    const f = field(panel(page), /^Category$/);
    await f.locator('[data-quick-create="categories"]').click();
    const pop = f.locator('.eco-qc-pop');
    await expect(pop).toContainText('New category');
    await expect(pop).toContainText('For Product, in Persian.');
    const bg = await pop.evaluate((el) => getComputedStyle(el).backgroundColor);
    expect(bg).not.toBe('rgb(255, 255, 255)');
    await pop.locator('input').fill(k + ' Quick EN');
    await pop.locator('[data-qc-create]').click();
    await expect(pop).toBeHidden();
    await expect(f.locator('button[aria-haspopup="listbox"]')).toContainText(k + ' Quick EN');
    const r = await base.get('/categories?q=' + encodeURIComponent(k + ' Quick EN'));
    const it = r.items.find((x: any) => x.name?.fa === k + ' Quick EN');
    expect(it.key).toBe((k + '-quick-en').replace(/_/g, '-'));
    created.categories.push(it.id);
  });
});

test.describe('project settings on a phone', () => {
  test.use({ pageOpts: { signedIn: true, mobile: true } } as any);

  test('the window is a full-screen sheet with a sideways section scroller', async ({ page }) => {
    await openSettings(page, 'sms');
    const box = await psWindow(page).boundingBox();
    expect(box!.width).toBeGreaterThanOrEqual(388);
    expect(box!.x).toBeLessThanOrEqual(1);
    expect(box!.height).toBeGreaterThanOrEqual(800);
    const list = psWindow(page).locator('.ps-list');
    expect(await list.evaluate((el) => el.scrollWidth > el.clientWidth)).toBe(true);
    expect(await list.evaluate((el) => getComputedStyle(el).flexDirection)).toBe('row');
    await expect(psWindow(page).locator('[data-ps-close]')).toBeVisible();
    await noHorizontalScroll(page);
    const overflow = await page.locator('#project-settings-body').evaluate((el) => el.scrollWidth - el.clientWidth);
    expect(overflow).toBeLessThanOrEqual(1);
    await tab(page, 'tags').scrollIntoViewIfNeeded();
    await tab(page, 'tags').click();
    await expect(page).toHaveURL(/#\/project-settings\/tags$/);
    await settle(page);
    const t = root(page).locator('table.eco-table').first();
    await expect(t.locator('thead')).toBeHidden();
    await psWindow(page).locator('[data-ps-close]').click();
    await expect(psWindow(page)).toHaveCount(0);
  });

  test('quick-create popover fits the phone', async ({ page }) => {
    await open(page, '#/economy/products/new');
    await settlePanel(page);
    const f = field(panel(page), /^دسته‌بندی$/);
    await f.locator('[data-quick-create="categories"]').click();
    const pop = f.locator('.eco-qc-pop');
    await expect(pop).toBeVisible();
    const b = await pop.boundingBox();
    expect(b!.x).toBeGreaterThanOrEqual(0);
    expect(b!.x + b!.width).toBeLessThanOrEqual(390);
    await page.keyboard.press('Escape');
    await expect(pop).toBeHidden();
  });
});
