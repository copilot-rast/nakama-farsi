import { expect, test } from '../../lib/fixtures';
import type { Page } from '@playwright/test';
import { deleteUser, devicePlayer, uid, type Player } from '../../lib/api';
import { go, noHorizontalScroll, open, root, rows, search, settle } from '../../lib/console';

test.describe.configure({ mode: 'serial' });

const made: Player[] = [];
let key = '';

test.beforeAll(async () => {
  key = uid('e2etb').replace(/_/g, '');
  for (let i = 0; i < 23; i++) made.push(await devicePlayer(`e2e-tb-${key}-${i}`, `${key}_${String(i).padStart(2, '0')}`));
});

test.afterAll(async () => {
  for (const p of made) await deleteUser(p.id);
});

async function box(page: Page, sel: string) {
  const b = await root(page).locator(sel).first().boundingBox();
  if (!b) throw new Error('no box for ' + sel);
  return b;
}

test('players toolbar: one row of controls, count, filter chip with clear, pager at the bottom', async ({ page }) => {
  await open(page, '#/moderation/players');
  await search(page, key);
  const bar = root(page).locator('.eco-toolbar');
  await expect(bar).toHaveAttribute('role', 'search');
  await expect(bar.locator('[data-list-count]')).toHaveText('۲۳ مورد');
  await expect(bar.locator('[data-list-refresh]')).toHaveAttribute('aria-label', 'تازه‌سازی');
  await expect(bar).not.toContainText('موردی وجود ندارد');

  const s = await box(page, '.eco-toolbar .eco-search input');
  const f = await box(page, '.eco-toolbar .eco-filter');
  const r = await box(page, '[data-list-refresh]');
  expect(Math.round(s.height)).toBe(32);
  expect(Math.round(f.height)).toBe(32);
  expect(Math.round(r.height)).toBe(32);
  expect(Math.abs((s.y + s.height / 2) - (f.y + f.height / 2))).toBeLessThan(1.5);

  const weights = await root(page).locator('.eco-toolbar .eco-filter-label').evaluateAll((els) => els.map((e) => getComputedStyle(e).fontWeight));
  expect(new Set(weights).size).toBe(1);

  const pager = root(page).locator('.eco-pager');
  await expect(pager).toHaveCount(1);
  await expect(pager).toHaveAttribute('aria-label', 'صفحه‌بندی');
  await expect(pager.locator('.eco-pager-range')).toHaveText('۱ تا ۲۰ از ۲۳');
  await expect(pager.locator('.eco-pager-page')).toHaveText('صفحهٔ ۱ از ۲');
  await expect(pager.locator('[data-page="first"]')).toBeDisabled();
  await expect(pager.locator('[data-page="prev"]')).toBeDisabled();
  const t = await box(page, 'table');
  const p = await box(page, '.eco-pager');
  expect(p.y).toBeGreaterThan(t.y + t.height - 1);
  expect(Math.round((await box(page, '.eco-pager [data-page="next"]')).height)).toBe(32);

  await pager.locator('[data-page="last"]').click();
  await settle(page);
  await expect(pager.locator('.eco-pager-range')).toHaveText('۲۱ تا ۲۳ از ۲۳');
  await expect(rows(page)).toHaveCount(3);
  expect(page.url()).toContain('page=2');
  await expect(pager.locator('[data-page="next"]')).toBeDisabled();
  await pager.locator('[data-page="first"]').click();
  await settle(page);
  await expect(rows(page)).toHaveCount(20);

  const clear = root(page).locator('[data-filters-clear]');
  await expect(clear).toBeHidden();
  const status = root(page).locator('.eco-filter[data-filter="status"]');
  await status.locator('select').selectOption('suspended');
  await settle(page);
  await expect(status).toHaveClass(/eco-filter-on/);
  await expect(clear).toBeVisible();
  expect(page.url()).toContain('status=suspended');
  await expect(root(page).locator('.eco-pager')).toBeHidden();
  await expect(bar.locator('[data-list-count]')).toBeHidden();
  const empty = root(page).getByRole('button', { name: 'پاک کردن فیلترها' }).last();
  await expect(empty).toBeVisible();
  await clear.click();
  await settle(page);
  await expect(status).not.toHaveClass(/eco-filter-on/);
  await expect(status.locator('select')).toHaveValue('');
  await expect(clear).toBeHidden();
  expect(page.url()).not.toContain('status=');
  await expect(rows(page)).toHaveCount(20);
  await expect(root(page).locator('.eco-search input')).toHaveValue(key);

  await root(page).locator('.eco-search input').press('Escape');
  await settle(page);
  await expect(root(page).locator('.eco-search input')).toHaveValue('');
  await expect.poll(() => page.url()).not.toContain('q=');
});

test('selection bar appears only with a selection; header checkbox selects the page', async ({ page }) => {
  await open(page, '#/moderation/players?q=' + key);
  await settle(page);
  const sel = root(page).locator('[data-bulk-bar]');
  await expect(sel).toBeHidden();
  await expect(root(page)).not.toContainText('بازیکنی انتخاب نشده');

  await root(page).locator(`[data-select="${made[3].id}"]`).click();
  await expect(sel).toBeVisible();
  await expect(sel).toContainText('۱ بازیکن انتخاب شده');
  await expect(root(page).locator(`tr:has([data-select="${made[3].id}"])`)).toHaveClass(/eco-row-on/);
  const head = root(page).locator('[data-select-all]');
  await expect(head).toHaveAttribute('aria-checked', 'mixed');
  await expect(head).toHaveAttribute('aria-label', 'انتخاب همهٔ ردیف‌های این صفحه');

  await head.click();
  await expect(sel).toContainText('۲۰ بازیکن انتخاب شده');
  await expect(head).toHaveAttribute('aria-checked', 'true');
  await expect(sel.getByRole('button', { name: 'انتخاب همهٔ این صفحه' })).toBeHidden();

  await sel.locator('[data-sel-clear]').click();
  await expect(sel).toBeHidden();
  await expect(head).toHaveAttribute('aria-checked', 'false');
  await expect(root(page).locator('tr.eco-row-on')).toHaveCount(0);

  await root(page).locator(`[data-select="${made[5].id}"]`).focus();
  await page.keyboard.press(' ');
  await expect(sel).toContainText('۱ بازیکن انتخاب شده');
});

test('template bundle pager and toolbar (cups) use the same controls', async ({ page }) => {
  await page.goto('/#/cups/list');
  const bar = page.locator('.cu-filters');
  await bar.waitFor({ timeout: 20_000 });
  await expect(bar).toHaveClass(/eco-toolbar/);
  const f = await bar.locator('.eco-tbf').first().boundingBox();
  expect(Math.round(f!.height)).toBe(32);
  await expect(bar.locator('.eco-tbf .eco-tf-hint')).toBeHidden();
});

test('form hints sit 6px under the input with room between lines', async ({ page }) => {
  await open(page, '#/moderation/players');
  await root(page).locator('[data-player-new]').click();
  const form = page.locator('[data-player-form]');
  const fieldBox = form.locator('.eco-field').filter({ has: page.locator('.eco-field-foot p:not([hidden])') }).first();
  const gap = await fieldBox.evaluate((el) => {
    const input = el.querySelector('input,select,textarea,button')!.getBoundingClientRect();
    const foot = el.querySelector('.eco-field-foot')!;
    const hint = foot.querySelector('p:not([hidden])')!;
    const cs = getComputedStyle(hint);
    return { gap: hint.getBoundingClientRect().top - input.bottom, lh: parseFloat(cs.lineHeight) / parseFloat(cs.fontSize) };
  });
  expect(gap.gap).toBeGreaterThanOrEqual(5.5);
  expect(gap.gap).toBeLessThanOrEqual(8);
  expect(gap.lh).toBeGreaterThanOrEqual(1.5);
});

test('english, light and 390px: toolbar wraps without sideways scroll, Latin digits', async ({ spawn }) => {
  const page = await spawn({ signedIn: true, locale: 'en', mobile: true, width: 390, height: 844 });
  await page.goto('/#/moderation/players?q=' + key);
  await page.locator('#economy-root .eco-toolbar').waitFor();
  await settle(page);
  await expect(root(page).locator('[data-list-count]')).toHaveText('23 items');
  await expect(root(page).locator('.eco-pager-page')).toHaveText('Page 1 of 2');
  await noHorizontalScroll(page);
  const s = await box(page, '.eco-toolbar .eco-search input');
  const r = await box(page, '[data-list-refresh]');
  expect(Math.abs((s.y + s.height / 2) - (r.y + r.height / 2))).toBeLessThan(2);
  await go(page, '#/economy/bots');
  await noHorizontalScroll(page);
});
