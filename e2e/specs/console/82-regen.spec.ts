import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid } from '../../lib/api';
import { cg } from '../../lib/campaign';
import { sql } from '../../lib/db';
import { field, open, panel, pickRef, root, rows, savePanel, search, settlePanel } from '../../lib/console';
import { CONSOLE_URL } from '../../lib/env';

const k = uid().replace(/_/g, '');
const ids: Record<string, string> = {};
const users: string[] = [];
let rawSettings = '';

test.describe.configure({ mode: 'serial' });

test.beforeAll(async () => {
  rawSettings = sql(`SELECT value::text FROM campaign_settings WHERE id=1`).trim();
  ids.gem = (await eco.post('/currencies', { key: k + 'g', name: { fa: 'الماس ' + k, en: 'Gem ' + k }, type: 'hard', defaultValue: 50, iconUrl: '', tagIds: [], metadata: {}, sort: 98 })).id;
  ids.heart = (await eco.post('/currencies', { key: k + 'h', name: { fa: 'جان ' + k, en: 'Heart ' + k }, type: 'soft', defaultValue: 5, iconUrl: '', tagIds: [], metadata: {}, sort: 99 })).id;
  ids.tag = (await base.post('/tags', { keyType: 'user', key: k + 'u', name: { fa: 'جان بی‌نهایت ' + k } })).id;
});

test.afterAll(async () => {
  for (const id of users) await deleteUser(id);
  if (rawSettings) sql(`UPDATE campaign_settings SET value='${rawSettings.replace(/'/g, "''")}'::jsonb WHERE id=1`);
  await cg.raw('GET', '/settings');
  for (const r of ['heart', 'gem']) if (ids[r]) await eco.del('/currencies/' + ids[r]);
  if (ids.tag) await base.del('/tags/' + ids.tag);
});

test.describe('regenerating currency', () => {
  test('turn regen on in the currency editor and read it back', async ({ page }) => {
    await open(page, '#/economy/currencies');
    await search(page, k + 'h');
    await expect(rows(page)).toHaveCount(1);
    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    const p = panel(page);
    await expect(field(p, /^سقف پرشدن$/)).toBeHidden();
    await p.locator('.x-toggle', { hasText: 'با گذشت زمان خودش پر شود' }).click();
    await field(p, /^سقف پرشدن$/).locator('input').fill('۵');
    await field(p, /^هر چند ثانیه یک واحد$/).locator('input').fill('900');
    await expect(p.locator('[data-regen-preview]')).toContainText('۱۵ دقیقه');
    await expect(p.locator('[data-regen-preview]')).toContainText('۷۵ دقیقه');
    await pickRef(p, /^پر کردن کامل با$/, k + 'g');
    await field(p, /^قیمت پر کردن کامل$/).locator('input').fill('۲۰');
    await pickRef(p, /^تگ «بی‌نهایت»$/, k + 'u');
    await savePanel(page);
    const saved = await eco.get('/currencies/' + ids.heart);
    expect(saved.regen).toMatchObject({ enabled: true, max: 5, intervalSeconds: 900, clamp: false, refillPrice: { currencyKey: k + 'g', amount: 20 }, unlimitedTag: k + 'u' });
    await expect(rows(page).first()).toContainText('تا ۵');
    const filtered = await eco.get('/currencies?regen=true&q=' + k);
    expect(filtered.items.map((x: any) => x.key)).toEqual([k + 'h']);
  });

  test('a player reads the regen state and refills with gems', async () => {
    const pl = await devicePlayer('e2e-regen-' + k);
    users.push(pl.id);
    const st = await rpc(pl, 'economy_regen_state', { currencies: [k + 'h'] });
    expect(st.status).toBe(200);
    expect(st.data.currencies[0]).toMatchObject({ currencyKey: k + 'h', value: 5, max: 5, intervalSeconds: 900, regen: true, nextAt: null, refillPrice: { currencyKey: k + 'g', amount: 20 } });
    const full = await rpc(pl, 'economy_regen_refill', { currencyKey: k + 'h', requestId: 'full-' + k });
    expect(full.status).not.toBe(200);
    sql(`UPDATE users SET wallet = jsonb_set(wallet, '{${k}h}', '1') WHERE id='${pl.id}'`);
    const one = await rpc(pl, 'economy_regen_state', { currencies: [k + 'h'] });
    expect(one.data.currencies[0].value).toBe(1);
    expect(one.data.currencies[0].nextAt).not.toBeNull();
    const refill = await rpc(pl, 'economy_regen_refill', { currencyKey: k + 'h', requestId: 'r-' + k });
    expect(refill.status).toBe(200);
    expect(refill.data).toMatchObject({ filled: 4, paid: { currencyKey: k + 'g', amount: 20 }, replayed: false });
    expect(refill.data.wallet[k + 'g']).toBe(30);
    const again = await rpc(pl, 'economy_regen_refill', { currencyKey: k + 'h', requestId: 'r-' + k });
    expect(again.data.replayed).toBe(true);
    const bal = await rpc(pl, 'economy_balances', {});
    expect(bal.data.wallet[k + 'h']).toBe(5);
    expect(bal.data.regen.find((x: any) => x.currencyKey === k + 'h')).toBeTruthy();
  });

  test('campaign settings show the play cost and save it', async ({ page }) => {
    await page.goto(CONSOLE_URL + '/#/campaign/settings');
    await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
    const r = page.locator('#campaign-root');
    await expect(r.getByText('در حال بارگذاری…')).toHaveCount(0);
    await expect(r.getByText('هزینهٔ بازی').first()).toBeVisible();
    await r.locator('select[name="play.costCurrency"]').selectOption(k + 'h');
    await r.locator('input[name="play.costAmount"]').fill('2');
    await r.locator('[data-campaign-save="settings"]').click();
    await expect.poll(async () => (await cg.raw('GET', '/settings')).data?.settings?.play?.costCurrency).toBe(k + 'h');
    const s = (await cg.raw('GET', '/settings')).data;
    expect(s.settings.play).toMatchObject({ costCurrency: k + 'h', costAmount: 2, refundOnWin: true });
    expect(s.costCurrencyRegen).toBe(true);
    await page.reload();
    await expect(r.locator('select[name="play.costCurrency"]')).toHaveValue(k + 'h');
  });

  test('regen off from the editor clears the setting', async ({ page }) => {
    await open(page, '#/economy/currencies');
    await search(page, k + 'h');
    await rows(page).first().locator('td').first().click();
    await settlePanel(page);
    const p = panel(page);
    await expect(field(p, /^سقف پرشدن$/).locator('input')).toHaveValue('۵');
    await p.locator('.x-toggle', { hasText: 'با گذشت زمان خودش پر شود' }).click();
    await savePanel(page);
    const saved = await eco.get('/currencies/' + ids.heart);
    expect(saved.regen).toBeNull();
    await expect(root(page)).toBeVisible();
  });
});
