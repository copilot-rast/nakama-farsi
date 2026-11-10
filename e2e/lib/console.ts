import { expect, type Locator, type Page } from '@playwright/test';
import { CONSOLE_PASSWORD, CONSOLE_URL, CONSOLE_USER } from './env';

export const ADMIN = { user: CONSOLE_USER, pass: CONSOLE_PASSWORD };

export async function loginAs(page: Page, who: { user: string; pass: string }) {
  await page.goto(CONSOLE_URL + '/#/login');
  const user = page.locator('input[name="username"]');
  await user.waitFor();
  await expect(async () => {
    await user.fill(who.user);
    await page.locator('input[name="password"]').fill(who.pass);
    await expect(user).toHaveValue(who.user);
  }).toPass();
  await page.keyboard.press('Enter');
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
}

export async function open(page: Page, hash: string) {
  await page.goto(CONSOLE_URL + '/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await root(page).waitFor({ timeout: 15_000 });
  await settle(page);
}

export async function go(page: Page, hash: string) {
  await page.evaluate((h) => { location.hash = h; }, hash);
  await root(page).waitFor();
  await settle(page);
}

export async function settle(page: Page, scope: Locator | string = ROOT) {
  const root = typeof scope === 'string' ? page.locator(scope) : scope;
  await expect(root.locator('.animate-pulse')).toHaveCount(0, { timeout: 15_000 });
}

export const ROOT = 'html[data-ps-open] #project-settings-body, html:not([data-ps-open]) #economy-root';

export function root(page: Page) {
  return page.locator(ROOT);
}

export function psWindow(page: Page) {
  return page.locator('[data-economy="project-settings"] .ps-window');
}

export async function openSettings(page: Page, section: string) {
  await open(page, '#/project-settings/' + section);
  await expect(psWindow(page)).toBeVisible();
  await expect(psWindow(page).locator(`[role="tab"][data-ps-section="${section}"]`)).toHaveAttribute('aria-selected', 'true');
}

export function panel(page: Page) {
  return page.locator('[data-economy="panel"]');
}

export async function settlePanel(page: Page) {
  await expect(panel(page).locator('form')).toBeVisible();
  await expect(panel(page).locator('.animate-pulse')).toHaveCount(0);
}

export function field(scope: Locator, label: string | RegExp) {
  return scope.locator('.eco-field').filter({ has: scope.page().locator('p > span', { hasText: label }) }).first();
}

export function translated(scope: Locator, name: string) {
  return scope.locator(`[data-translated="${name}"]`).first();
}

export async function fillTranslated(scope: Locator, name: string, values: Record<string, string>) {
  const box = translated(scope, name);
  for (const [code, text] of Object.entries(values)) {
    await box.locator(`[role="tab"][data-lang="${code}"]`).click();
    const ctl = box.locator(`[role="tabpanel"][data-lang="${code}"]`).locator('input,textarea').first();
    await ctl.fill(text);
  }
}

export async function pickRef(scope: Locator, label: string | RegExp, search: string, option?: string | RegExp) {
  const f = field(scope, label);
  await f.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = f.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(search);
  const opt = pop.locator('[role="option"]', { hasText: option || search }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

export async function savePanel(page: Page) {
  const p = panel(page);
  await p.locator('.x-form > div:last-child button.eco-btn').last().click();
  const bad = p.locator('.eco-invalid p.text-error-600:not([hidden]), [role="alert"], .eco-row-invalid .eco-items-err:not([hidden])');
  await expect.poll(async () => (await p.count()) === 0 || (await bad.count()) > 0, { timeout: 15_000 }).toBe(true);
  if (await p.count()) throw new Error('form not saved: ' + (await bad.allInnerTexts()).join(' | '));
  await expect(toasts(page).last()).toBeVisible();
}

export function toasts(page: Page) {
  return page.locator('.eco-toasts .eco-notice');
}

export function rows(page: Page) {
  return root(page).locator('tbody tr.x-table-row');
}

export async function search(page: Page, text: string) {
  const res = page.waitForResponse((r) => r.url().includes('q=' + encodeURIComponent(text)) && r.request().method() === 'GET');
  await root(page).locator('.eco-search input').fill(text);
  await res;
  await settle(page);
  await expect(root(page).locator('.opacity-60')).toHaveCount(0);
}

export async function noHorizontalScroll(page: Page) {
  const over = await page.evaluate(() => {
    const doc = document.documentElement;
    const out = [doc.scrollWidth - doc.clientWidth];
    const main = document.getElementById('main');
    if (main) out.push(main.scrollWidth - main.clientWidth);
    return Math.max(...out);
  });
  expect(over, 'page scrolls sideways by ' + over + 'px').toBeLessThanOrEqual(1);
}

export function watchErrors(page: Page) {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));
  return errors;
}
