import { expect, test } from '@playwright/test';
import { API_URL, CHROME } from '../../lib/env';

test.use({ launchOptions: { executablePath: CHROME } });

test('API docs page loads from the server only', async ({ page }) => {
  const origin = new URL(API_URL).origin;
  const outside: string[] = [];
  page.on('request', (r) => {
    const u = r.url();
    if (!u.startsWith(origin) && !u.startsWith('data:') && !u.startsWith('blob:')) outside.push(u);
  });
  const doc = await page.request.get(API_URL + '/docs/openapi.json');
  expect(doc.status()).toBe(200);
  const body = await doc.json();
  expect(body.openapi).toMatch(/^3\./);
  expect(Object.keys(body.paths)).toContain('/v2/rpc/base_account_get');
  await page.goto(API_URL + '/docs');
  await expect(page.locator('.swagger-ui .info .title')).toContainText('Nakama API');
  await expect(page.locator('.opblock-tag-section').first()).toBeVisible();
  await page.locator('.opblock-tag', { hasText: 'base' }).first().click();
  await expect(page.locator('.opblock-summary-path', { hasText: '/v2/rpc/base_account_get' })).toBeVisible();
  await page.waitForLoadState('networkidle');
  expect(outside).toEqual([]);
});
