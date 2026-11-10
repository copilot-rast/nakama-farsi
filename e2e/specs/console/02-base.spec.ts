import { expect, test } from '../../lib/fixtures';
import { base, eco, uid } from '../../lib/api';
import { field, fillTranslated, go, open, panel, pickRef, root, rows, savePanel, search, settle, settlePanel, toasts, translated } from '../../lib/console';

const created = { categories: [] as string[], tags: [] as string[], prizes: [] as string[], langs: [] as string[] };

test.afterAll(async () => {
  for (const id of created.prizes) await eco.del('/prizes/' + id);
  for (const id of created.categories) await base.del('/categories/' + id);
  for (const id of created.tags) await base.del('/tags/' + id);
  for (const code of created.langs) {
    const l = await base.get('/languages/' + code).catch(() => null);
    if (!l) continue;
    if (l.isActive) await base.put('/languages/' + code, { ...l, isActive: false });
    await base.del('/languages/' + code);
  }
});

async function findByKey(resource: 'categories' | 'tags', key: string) {
  const r = await base.get(`/${resource}?q=${encodeURIComponent(key)}&limit=5`);
  return r.items.find((x: any) => x.key === key);
}

test.describe('base: languages', () => {
  test('add a language, see it in forms, deactivate and delete it; the default stays locked', async ({ page }) => {
    const code = 'xa-' + Date.now().toString(36).slice(-6);
    created.langs.push(code);
    await open(page, '#/project-settings/languages');
    await root(page).getByRole('button', { name: 'زبان جدید' }).click();
    await settlePanel(page);
    const p = panel(page);
    const codeField = field(p, /^کد زبان$/);
    await codeField.locator('input').fill(code.toUpperCase());
    await field(p, /^نام بومی$/).locator('input').fill('Testish ' + code);
    await fillTranslated(p, 'name', { fa: 'زبان آزمایشی', en: 'Test language' });
    await field(p, /^جهت نوشتن$/).locator('select').selectOption('ltr');
    await p.getByRole('button', { name: 'ایجاد' }).click();
    await expect(codeField).toHaveClass(/eco-invalid/);
    await expect(codeField).toContainText('حروف کوچک');
    await codeField.locator('input').fill(code);
    await expect(codeField).not.toHaveClass(/eco-invalid/);
    await savePanel(page);
    await search(page, code);
    const row = rows(page).filter({ hasText: code });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText('Testish ' + code);
    await expect(row).toContainText('فعال');
    expect((await base.get('/languages/' + code)).isActive).toBe(true);

    await go(page, '#/project-settings/categories/new');
    await settlePanel(page);
    await expect(translated(panel(page), 'name').locator(`[role="tab"][data-lang="${code}"]`)).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);

    await go(page, '#/project-settings/languages?q=' + code);
    const row2 = rows(page).filter({ hasText: code });
    await row2.locator('.x-toggle').click();
    await expect(toasts(page).filter({ hasText: 'غیرفعال شد' })).toBeVisible();
    await expect(row2).toContainText('غیرفعال');
    await expect.poll(async () => (await base.get('/languages/' + code)).isActive).toBe(false);

    await go(page, '#/project-settings/categories/new');
    await settlePanel(page);
    await expect(translated(panel(page), 'name').locator('[role="tab"][data-lang="fa"]')).toBeVisible();
    await expect(translated(panel(page), 'name').locator(`[role="tab"][data-lang="${code}"]`)).toHaveCount(0);
    await page.keyboard.press('Escape');

    await go(page, '#/project-settings/languages');
    await search(page, 'fa');
    const fa = rows(page).filter({ has: page.locator('span.font-mono', { hasText: /^fa$/ }) });
    await expect(fa).toContainText('پیش‌فرض');
    await fa.locator('.x-toggle').click();
    await expect(toasts(page).filter({ hasText: 'زبان پیش‌فرض حذف یا غیرفعال نمی‌شود' })).toBeVisible();
    await expect(fa.locator('input[role="switch"]')).toHaveAttribute('aria-checked', 'true');
    expect((await base.get('/languages/fa')).isActive).toBe(true);

    await search(page, code);
    await rows(page).filter({ hasText: code }).locator('td').first().click();
    await settlePanel(page);
    await expect(field(panel(page), /^کد زبان$/).locator('input')).toHaveValue(code);
    await expect(field(panel(page), /^کد زبان$/).locator('input')).toHaveJSProperty('readOnly', true);
    await panel(page).getByRole('button', { name: 'حذف' }).click();
    await panel(page).getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(panel(page)).toHaveCount(0);
    await expect(toasts(page).filter({ hasText: 'حذف شد' })).toBeVisible();
    await expect(root(page).getByText('با این جستجو چیزی پیدا نشد.')).toBeVisible();
    expect((await base.del('/languages/' + code)).status).toBe(404);
  });

  test('coverage panel numbers move after a translation is filled in', async ({ page }) => {
    const key = uid();
    const cat = await base.post('/categories', { key, keyType: 'news', name: { fa: 'پوشش ' + key }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    created.categories.push(cat.id);
    await open(page, '#/project-settings/languages');
    const cov = root(page).locator('[data-coverage]');
    await expect(cov.locator('[data-cov-lang="en"]')).toBeVisible();
    const before = Number(await cov.locator('[data-cov-lang="en"]').getAttribute('data-cov-missing'));
    expect(before).toBeGreaterThan(0);
    await expect(cov.locator('[data-cov="categories:en"]')).toContainText('ناقص');

    await go(page, '#/project-settings/categories/' + cat.id);
    await settlePanel(page);
    await fillTranslated(panel(page), 'name', { en: 'Coverage ' + key });
    await savePanel(page);

    await go(page, '#/project-settings/languages');
    await expect(cov.locator('[data-cov-lang="en"]')).toHaveAttribute('data-cov-missing', String(before - 1));
    const reload = page.waitForResponse((r) => r.url().includes('/languages/coverage'));
    await cov.getByRole('button', { name: 'تازه‌سازی' }).click();
    await reload;
    await settle(page);
    await expect(cov.locator('[data-cov-lang="en"]')).toHaveAttribute('data-cov-missing', String(before - 1));
  });
});

test.describe('base: categories and tags', () => {
  test('category CRUD with a key type, filter and an IN_USE delete', async ({ page }) => {
    const key = uid();
    await open(page, '#/project-settings/categories');
    await root(page).getByRole('button', { name: 'دسته‌بندی جدید' }).first().click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill(key);
    await field(p, /^نوع کلید$/).locator('select').selectOption('prize');
    await fillTranslated(p, 'name', { fa: 'جایزه‌های ' + key, en: 'Prizes ' + key });
    await fillTranslated(p, 'description', { fa: 'توضیح آزمایشی' });
    await savePanel(page);
    const cat = await findByKey('categories', key);
    expect(cat).toBeTruthy();
    created.categories.push(cat.id);
    expect(cat.keyType).toBe('prize');
    expect(cat.name).toEqual({ fa: 'جایزه‌های ' + key, en: 'Prizes ' + key });
    expect(cat.description).toEqual({ fa: 'توضیح آزمایشی' });

    await search(page, key);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('جایزه‌های ' + key);
    await expect(rows(page).first()).toContainText('جایزه');
    const typeFilter = root(page).locator('.eco-filters select').first();
    await typeFilter.selectOption('product');
    await settle(page);
    await expect(root(page).getByText('با این جستجو چیزی پیدا نشد.')).toBeVisible();
    await expect(page).toHaveURL(/keyType=product/);
    await typeFilter.selectOption('prize');
    await expect(rows(page)).toHaveCount(1);

    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    await fillTranslated(panel(page), 'name', { fa: 'ویرایش‌شده ' + key });
    await savePanel(page);
    await expect(rows(page).first()).toContainText('ویرایش‌شده ' + key);

    const prize = await eco.post('/prizes', { key: uid(), name: { fa: 'جایزه ' + key }, message: {}, categoryId: cat.id, tagIds: [], isRepeatable: false, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }], metadata: {} });
    created.prizes.push(prize.id);
    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    await panel(page).getByRole('button', { name: 'حذف' }).click();
    await panel(page).getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(panel(page).getByRole('alert')).toContainText('این مورد جای دیگری استفاده شده و حذف نمی‌شود.');
    await page.keyboard.press('Escape');
    await eco.del('/prizes/' + prize.id);
    created.prizes.splice(created.prizes.indexOf(prize.id), 1);
    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    await panel(page).getByRole('button', { name: 'حذف' }).click();
    await panel(page).getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(panel(page)).toHaveCount(0);
    await expect(root(page).getByText('با این جستجو چیزی پیدا نشد.')).toBeVisible();
    expect(await findByKey('categories', key)).toBeUndefined();
    created.categories.splice(created.categories.indexOf(cat.id), 1);
  });

  test('tag CRUD and pickers filtered by key type (+ general for tags)', async ({ page }) => {
    const k = uid();
    const prizeCat = await base.post('/categories', { key: k + '_pc', keyType: 'prize', name: { fa: 'دسته جایزه ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    const productCat = await base.post('/categories', { key: k + '_dc', keyType: 'product', name: { fa: 'دسته محصول ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    created.categories.push(prizeCat.id, productCat.id);
    const productTag = await base.post('/tags', { key: k + '_dt', keyType: 'product', name: { fa: 'تگ محصول ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    const generalTag = await base.post('/tags', { key: k + '_gt', keyType: 'general', name: { fa: 'تگ عمومی ' + k }, description: {}, iconUrl: '', metadata: {}, sort: 0 });
    created.tags.push(productTag.id, generalTag.id);

    await open(page, '#/project-settings/tags');
    await root(page).getByRole('button', { name: 'تگ جدید' }).first().click();
    await settlePanel(page);
    await expect(field(panel(page), /^نوع کلید$/).locator('select')).toHaveValue('general');
    await field(panel(page), /^کلید$/).locator('input').fill(k + '_pt');
    await field(panel(page), /^نوع کلید$/).locator('select').selectOption('prize');
    await fillTranslated(panel(page), 'name', { fa: 'تگ جایزه ' + k });
    await savePanel(page);
    const prizeTag = await findByKey('tags', k + '_pt');
    created.tags.push(prizeTag.id);
    expect(prizeTag.keyType).toBe('prize');
    await search(page, k);
    await expect(rows(page)).toHaveCount(3);
    await root(page).locator('.eco-filters select').first().selectOption('general');
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('تگ عمومی ' + k);

    await go(page, '#/economy/prizes/new');
    await settlePanel(page);
    const p = panel(page);
    const catField = field(p, /^دسته‌بندی$/);
    await catField.locator('button[aria-haspopup="listbox"]').click();
    await catField.locator('.eco-pop input').fill(k);
    await expect(catField.locator('[role="option"]')).toHaveCount(1);
    await expect(catField.locator('[role="option"]').first()).toContainText('دسته جایزه ' + k);
    await page.keyboard.press('Escape');
    const tagField = field(p, /^تگ‌ها$/);
    await tagField.locator('button[aria-haspopup="listbox"]').click();
    await tagField.locator('.eco-pop input').fill(k);
    await expect(tagField.locator('[role="option"]')).toHaveCount(2);
    await expect(tagField.locator('[role="option"]', { hasText: 'تگ جایزه ' + k })).toBeVisible();
    await expect(tagField.locator('[role="option"]', { hasText: 'تگ عمومی ' + k })).toBeVisible();
    await expect(tagField.locator('[role="option"]', { hasText: 'تگ محصول ' + k })).toHaveCount(0);
    await tagField.locator('[role="option"]', { hasText: 'تگ جایزه ' + k }).click();
    await expect(tagField.locator('.eco-tag', { hasText: 'تگ جایزه ' + k })).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);

    await go(page, '#/economy/products/new');
    await settlePanel(page);
    await pickRef(panel(page), /^دسته‌بندی$/, k, 'دسته محصول ' + k);
    const ptags = field(panel(page), /^تگ‌ها$/);
    await ptags.locator('button[aria-haspopup="listbox"]').click();
    await ptags.locator('.eco-pop input').fill(k);
    await expect(ptags.locator('[role="option"]')).toHaveCount(2);
    await expect(ptags.locator('[role="option"]', { hasText: 'تگ محصول ' + k })).toBeVisible();
    await expect(ptags.locator('[role="option"]', { hasText: 'تگ جایزه ' + k })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await page.keyboard.press('Escape');

    await go(page, '#/project-settings/tags?q=' + k + '_pt');
    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    await panel(page).getByRole('button', { name: 'حذف' }).click();
    await panel(page).getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(panel(page)).toHaveCount(0);
    await expect(root(page).getByText('با این جستجو چیزی پیدا نشد.')).toBeVisible();
    created.tags.splice(created.tags.indexOf(prizeTag.id), 1);
  });
});
