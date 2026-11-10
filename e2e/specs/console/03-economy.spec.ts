import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { field, fillTranslated, go, open, panel, pickRef, root, rows, savePanel, search, settle, settlePanel, toasts } from '../../lib/console';

const k = uid();
const ids: Record<string, string> = {};
const players: string[] = [];
const taxo: { categories: string[]; tags: string[] } = { categories: [], tags: [] };

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const r of ['gifts:reward', 'gifts:discount', 'products', 'prizes', 'metas', 'currencies']) {
    const id = ids[r];
    if (id) await eco.del('/' + r.split(':')[0] + '/' + id);
  }
  for (const p of players) await deleteUser(p);
  for (const id of taxo.tags) await base.del('/tags/' + id);
  for (const id of taxo.categories) await base.del('/categories/' + id);
});

async function quickCreate(page: import('@playwright/test').Page, label: RegExp, resource: 'categories' | 'tags', name: string) {
  const f = field(panel(page), label);
  await f.locator(`[data-quick-create="${resource}"]`).click();
  const pop = f.locator('.eco-qc-pop');
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(name);
  await pop.locator('[data-qc-create]').click();
  await expect(pop).toBeHidden();
  const r = await base.get(`/${resource}?q=${encodeURIComponent(name)}&limit=10`);
  const item = r.items.find((x: any) => x.name?.fa === name);
  expect(item).toBeTruthy();
  taxo[resource].push(item.id);
  return item;
}

async function byKey(resource: string, key: string, prop = 'key') {
  const r = await eco.get(`/${resource}?q=${encodeURIComponent(key)}&limit=10`);
  return r.items.find((x: any) => x[prop] === key);
}

test.describe('economy', () => {
  test('currency: create with Persian digits, edit, list shows it', async ({ page }) => {
    await open(page, '#/economy/currencies');
    await root(page).getByRole('button', { name: 'ارز جدید' }).first().click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill(k + 'c');
    await field(p, /^نوع$/).locator('select').selectOption('hard');
    await fillTranslated(p, 'name', { fa: 'الماس ' + k, en: 'Gem ' + k });
    await field(p, /^مقدار اولیه$/).locator('input').fill('۲۵');
    await savePanel(page);
    const c = await byKey('currencies', k + 'c');
    ids.currencies = c.id;
    expect(c).toMatchObject({ type: 'hard', defaultValue: 25, name: { fa: 'الماس ' + k, en: 'Gem ' + k } });
    await search(page, k + 'c');
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('الماس ' + k);
    await expect(rows(page).first()).toContainText('سخت');
    await expect(rows(page).first()).toContainText('۲۵');
    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    await expect(field(panel(page), /^کلید$/).locator('input')).toHaveJSProperty('readOnly', true);
    await expect(field(panel(page), /^مقدار اولیه$/).locator('input')).toHaveValue('۲۵');
    await field(panel(page), /^ترتیب$/).locator('input').fill('7');
    await savePanel(page);
    await expect(rows(page).first()).toContainText('۷');
  });

  test('meta: server-only meta with a decimal default', async ({ page }) => {
    await open(page, '#/economy/metas');
    await root(page).getByRole('button', { name: 'متای جدید' }).first().click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill(k + 'm');
    await fillTranslated(p, 'name', { fa: 'سطح ' + k });
    await field(p, /^مقدار اولیه$/).locator('input').fill('1.5');
    await p.locator('.x-toggle', { hasText: 'فقط سرور' }).click();
    await savePanel(page);
    const m = await byKey('metas', k + 'm');
    ids.metas = m.id;
    expect(m).toMatchObject({ defaultValue: 1.5, serverOnly: true });
    await search(page, k + 'm');
    await expect(rows(page).first()).toContainText('فقط سرور');
  });

  test('prize: item rows for currency, meta and storage', async ({ page }) => {
    await open(page, '#/economy/prizes/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill(k + 'p');
    await fillTranslated(p, 'name', { fa: 'بستهٔ ' + k, en: 'Bundle ' + k });
    await fillTranslated(p, 'message', { fa: 'تبریک!' });
    const itemRows = p.locator('.eco-items-row');
    await expect(itemRows).toHaveCount(1);
    await itemRows.nth(0).locator('button[aria-haspopup="listbox"]').click();
    await itemRows.nth(0).locator('.eco-pop input').fill(k + 'c');
    await itemRows.nth(0).locator('[role="option"]', { hasText: k + 'c' }).click();
    await itemRows.nth(0).locator('input[aria-label="مقدار"]').fill('۱۰');
    await p.getByRole('button', { name: 'افزودن ردیف' }).click();
    await expect(itemRows).toHaveCount(2);
    await itemRows.nth(1).locator('select[aria-label="نوع"]').selectOption('meta');
    await itemRows.nth(1).locator('button[aria-haspopup="listbox"]').click();
    await itemRows.nth(1).locator('.eco-pop input').fill(k + 'm');
    await itemRows.nth(1).locator('[role="option"]', { hasText: k + 'm' }).click();
    await itemRows.nth(1).locator('select[aria-label="عمل"]').selectOption('set');
    await itemRows.nth(1).locator('input[aria-label="مقدار"]').fill('3');
    await p.getByRole('button', { name: 'افزودن ردیف' }).click();
    await itemRows.nth(2).locator('select[aria-label="نوع"]').selectOption('storage');
    await itemRows.nth(2).locator('input[aria-label="کلید"]').fill('e2e_badge');
    await itemRows.nth(2).locator('input[aria-label="مقدار"]').fill('{"tier":');
    await p.getByRole('button', { name: 'ایجاد' }).click();
    await expect(p.locator('.eco-row-invalid')).toHaveCount(1);
    await itemRows.nth(2).locator('input[aria-label="مقدار"]').fill('{"tier":2}');
    await savePanel(page);
    const prize = await byKey('prizes', k + 'p');
    ids.prizes = prize.id;
    expect(prize.items).toEqual([
      { type: 'currency', key: k + 'c', action: 'add', value: 10 },
      { type: 'meta', key: k + 'm', action: 'set', value: 3 },
      { type: 'storage', key: 'e2e_badge', action: 'give', value: { tier: 2 } },
    ]);
    expect(prize.message).toEqual({ fa: 'تبریک!' });
    await go(page, '#/economy/prizes?q=' + k + 'p');
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('+۱۰ الماس ' + k);
    await expect(rows(page).first()).toContainText('e2e_badge');
    await expect(rows(page).first()).toContainText('یک‌باره');
  });

  test('product: fa/en/ar names, methods drive the fields, sku and coin price', async ({ page }) => {
    await open(page, '#/economy/products/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill(k + 'd');
    await fillTranslated(p, 'name', { fa: 'محصول ' + k, en: 'Product ' + k, ar: 'منتج ' + k });
    await fillTranslated(p, 'description', { fa: 'شرح', en: 'About' });
    await pickRef(p, /^جایزه$/, k + 'p');
    const method = (name: string) => p.locator('.x-checkbox', { hasText: name });
    await expect(method('زرین‌پال')).toHaveAttribute('aria-checked', 'true');
    await expect(field(p, /^SKU بازار$/)).toBeHidden();
    await expect(field(p, /^قیمت با سکه$|^ارز$/)).toBeHidden();
    await method('بازار').click();
    await method('مایکت').click();
    await method('سکهٔ بازی').click();
    await expect(field(p, /^SKU بازار$/)).toBeVisible();
    await expect(field(p, /^SKU مایکت$/)).toBeVisible();
    await field(p, /^قیمت \(تومان\)$/).locator('input').fill('50');
    await p.getByRole('button', { name: 'ایجاد' }).click();
    await expect(field(p, /^SKU بازار$/)).toContainText('این فیلد لازم است.');
    await expect(field(p, /^قیمت \(تومان\)$/)).toHaveClass(/eco-invalid/);
    await field(p, /^قیمت \(تومان\)$/).locator('input').fill('۱۲۰۰۰');
    await field(p, /^SKU بازار$/).locator('input').fill(k + '_bz');
    await field(p, /^SKU مایکت$/).locator('input').fill(k + '_mk');
    await pickRef(p, /^ارز$/, k + 'c');
    await field(p, /^مقدار$/).locator('input').fill('5');
    await savePanel(page);
    const prod = await byKey('products', k + 'd');
    ids.products = prod.id;
    expect(prod).toMatchObject({
      name: { fa: 'محصول ' + k, en: 'Product ' + k, ar: 'منتج ' + k },
      paymentMethods: ['zarinpal', 'bazaar', 'myket', 'in_app'],
      priceToman: 12000,
      sku: { bazaar: k + '_bz', myket: k + '_mk' },
      coinPrice: { currencyKey: k + 'c', amount: 5 },
      prizeId: ids.prizes,
      isActive: true,
    });
    await go(page, '#/economy/products?q=' + k + 'd');
    const row = rows(page).first();
    await expect(row).toContainText('محصول ' + k);
    await expect(row).toContainText('۱۲٬۰۰۰ تومان');
    for (const m of ['زرین‌پال', 'بازار', 'مایکت', 'سکهٔ بازی']) await expect(row).toContainText(m);
    await root(page).locator('.eco-filters select').first().selectOption('false');
    await expect(root(page).getByText('با این جستجو چیزی پیدا نشد.')).toBeVisible();
  });

  test('gift codes: a reward code and a discount code', async ({ page }) => {
    const code = ('E2E' + k.slice(4, 14)).toUpperCase();
    await open(page, '#/economy/gifts/new');
    await settlePanel(page);
    let p = panel(page);
    await fillTranslated(p, 'name', { fa: 'هدیهٔ ' + k });
    await field(p, /^کد$/).locator('input').fill(code.toLowerCase() + 'r');
    await expect(field(p, /^کد$/).locator('input')).toHaveCSS('text-transform', 'uppercase');
    await pickRef(p, /^جایزه$/, k + 'p');
    await savePanel(page);
    const reward = await byKey('gifts', code + 'R', 'code');
    ids['gifts:reward'] = reward.id;
    expect(reward).toMatchObject({ type: 'reward', prizeId: ids.prizes, perUserLimit: 1, isActive: true });

    await go(page, '#/economy/gifts/new');
    await settlePanel(page);
    p = panel(page);
    await fillTranslated(p, 'name', { fa: 'تخفیف ' + k });
    await field(p, /^نوع$/).locator('select').selectOption('discount');
    await expect(field(p, /^جایزه$/)).toBeHidden();
    await p.getByRole('button', { name: 'ایجاد' }).click();
    await expect(field(p, /^کد$/)).toHaveClass(/eco-invalid/);
    await expect(field(p, /^درصد تخفیف$/)).toHaveClass(/eco-invalid/);
    await field(p, /^کد$/).locator('input').fill(code + 'D');
    await field(p, /^درصد تخفیف$/).locator('input').fill('30');
    await savePanel(page);
    const disc = await byKey('gifts', code + 'D', 'code');
    ids['gifts:discount'] = disc.id;
    expect(disc).toMatchObject({ type: 'discount', discountPercent: 30 });

    await go(page, '#/economy/gifts?q=' + code);
    await expect(rows(page)).toHaveCount(2);
    const rRow = rows(page).filter({ hasText: code + 'R' });
    const dRow = rows(page).filter({ hasText: code + 'D' });
    await expect(rRow).toContainText('جایزه');
    await expect(dRow).toContainText('تخفیف');
    await expect(dRow).toContainText('۳۰٪');
  });

  test('currency and gift: category and tag pickers by key type, quick-create, list column and catalog', async ({ page }) => {
    const curCat = await base.post('/categories', { key: k + '_cc', keyType: 'currency', name: { fa: 'ارز ویژه ' + k } });
    const prodCat = await base.post('/categories', { key: k + '_pc', keyType: 'product', name: { fa: 'ارز محصول ' + k } });
    const general = await base.post('/tags', { key: k + '_gt', keyType: 'general', name: { fa: 'همه‌جا ' + k } });
    taxo.categories.push(curCat.id, prodCat.id);
    taxo.tags.push(general.id);

    await open(page, '#/economy/currencies/' + ids.currencies);
    await settlePanel(page);
    let p = panel(page);
    const catField = field(p, /^دسته‌بندی$/);
    await catField.locator('button[aria-haspopup="listbox"]').first().click();
    const pop = catField.locator('.eco-pop').first();
    await pop.locator('input').fill('ارز');
    await expect(pop.locator('[role="option"]', { hasText: 'ارز ویژه ' + k })).toBeVisible();
    await expect(pop.locator('[role="option"]', { hasText: 'ارز محصول ' + k })).toHaveCount(0);
    await pop.locator('[role="option"]', { hasText: 'ارز ویژه ' + k }).click();
    const tagName = 'تگ ارز ' + k;
    const curTag = await quickCreate(page, /^تگ‌ها$/, 'tags', tagName);
    expect(curTag.keyType).toBe('currency');
    await savePanel(page);
    const c = await eco.get('/currencies/' + ids.currencies);
    expect(c.categoryId).toBe(curCat.id);
    expect(c.tagIds).toEqual([curTag.id]);
    await search(page, k + 'c');
    await expect(rows(page).first().locator('[data-taxonomy]')).toContainText('ارز ویژه ' + k);
    await expect(rows(page).first().locator('[data-taxonomy]')).toContainText(tagName);

    const player = await devicePlayer('e2e-tax-' + k);
    players.push(player.id);
    const cat = await rpc(player, 'economy_catalog', { lang: 'fa' });
    const cur = cat.data.currencies.find((x: any) => x.key === k + 'c');
    expect(cur).toMatchObject({ categoryKey: k + '_cc', tags: [curTag.key] });

    await open(page, '#/economy/gifts/' + ids['gifts:reward']);
    await settlePanel(page);
    p = panel(page);
    const giftCat = await quickCreate(page, /^دسته‌بندی$/, 'categories', 'دسته هدیه ' + k);
    expect(giftCat.keyType).toBe('gift');
    const giftTag = await quickCreate(page, /^تگ‌ها$/, 'tags', 'تگ هدیه ' + k);
    expect(giftTag.keyType).toBe('gift');
    const tagsField = field(p, /^تگ‌ها$/);
    await tagsField.locator('button[aria-haspopup="listbox"]').first().click();
    const tpop = tagsField.locator('.eco-pop').first();
    await tpop.locator('input').fill(k);
    await expect(tpop.locator('[role="option"]', { hasText: 'همه‌جا ' + k })).toBeVisible();
    await expect(tpop.locator('[role="option"]', { hasText: tagName })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await savePanel(page);
    const g = await eco.get('/gifts/' + ids['gifts:reward']);
    expect(g.categoryId).toBe(giftCat.id);
    expect(g.tagIds).toEqual([giftTag.id]);
    await go(page, '#/economy/gifts?categoryId=' + giftCat.id);
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first().locator('[data-taxonomy]')).toContainText('دسته هدیه ' + k);
    await expect(rows(page).first().locator('[data-taxonomy]')).toContainText('تگ هدیه ' + k);
  });

  test('payment methods page masks secrets and keeps them on save', async ({ page }) => {
    try {
      await open(page, '#/project-settings/payment-methods');
      const bazaar = root(page).locator('[data-method="bazaar"]');
      await expect(bazaar).toBeVisible();
      const secret = field(bazaar, /^کلید API$/);
      await expect(secret.locator('input')).toHaveValue('********');
      await expect(secret.locator('input')).toHaveClass(/eco-secret/);
      await expect(secret).toContainText('برای نگه‌داشتن مقدار فعلی، دست نزنید.');
      await expect(bazaar).toContainText('تنظیم شده');
      const before = (await eco.get('/payment-methods')).find((m: any) => m.method === 'bazaar');
      await field(bazaar, /^نام بسته$/).locator('input').fill('ir.e2e.' + k.slice(4));
      await bazaar.getByRole('button', { name: 'ذخیره' }).click();
      await expect(toasts(page).filter({ hasText: 'ذخیره شد.' })).toBeVisible();
      const after = (await eco.get('/payment-methods')).find((m: any) => m.method === 'bazaar');
      expect(after.config.packageName).toBe('ir.e2e.' + k.slice(4));
      expect(after.config.apiSecret).toBe(before.config.apiSecret);
      expect(after.configured).toBe(true);
      expect(after.config.sandbox).toBe(true);
      await page.reload();
      await settle(page);
      await expect(field(root(page).locator('[data-method="bazaar"]'), /^کلید API$/).locator('input')).toHaveValue('********');
      const html = await page.content();
      expect(html).not.toMatch(/apiSecret"\s*:\s*"(?!\*{8})[^"]+"/);
    } finally {
      restoreConfig();
    }
  });

  test('payments list and detail for a sandbox store purchase', async ({ page }) => {
    const player = await devicePlayer('e2e-pay-' + k);
    players.push(player.id);
    const cat = await rpc(player, 'economy_catalog', { lang: 'fa' });
    const pack = cat.data.products.find((x: any) => x.key === 'star_pack_small');
    const start = await rpc(player, 'economy_purchase_start', { productId: pack.id, method: 'bazaar' });
    expect(start.data.status).toBe('store');
    const token = `sandbox-${start.data.paymentId}-${k}`;
    const done = await rpc(player, 'economy_purchase_submit', { paymentId: start.data.paymentId, token });
    expect(done.data.status).toBe('success');

    await open(page, '#/economy/payments?userId=' + player.id);
    await expect(rows(page)).toHaveCount(1);
    const row = rows(page).first();
    await expect(row).toContainText(player.username);
    await expect(row).toContainText('star_pack_small');
    await expect(row).toContainText('بازار');
    await expect(row).toContainText('آزمایشی');
    await expect(row).toContainText('موفق');
    await row.locator('td').first().click();
    await expect(page).toHaveURL(new RegExp('#/economy/payments/' + start.data.paymentId));
    await settle(page);
    await expect(root(page).locator('h1')).toContainText(start.data.paymentId.slice(0, 8));
    await expect(root(page)).toContainText('داده شد');
    await expect(root(page)).toContainText(token);
    await expect(root(page)).toContainText('۱۰٬۰۰۰ تومان');
    const log = root(page).locator('.eco-log').first();
    await expect(log).toBeVisible();
    await log.locator('button[aria-expanded]').click();
    await expect(log.locator('.eco-log-detail')).toBeVisible();
    await root(page).getByRole('button', { name: 'بازگشت' }).click();
    await expect(page).toHaveURL(/#\/economy\/payments\?userId=/);
    await root(page).locator('.eco-filters select').first().selectOption('failed');
    await expect(root(page).getByText('با این جستجو چیزی پیدا نشد.')).toBeVisible();
  });
});
