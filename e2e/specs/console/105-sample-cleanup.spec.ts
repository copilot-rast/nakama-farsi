import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, deleteUser, devicePlayer, must, rpc, uid } from '../../lib/api';
import { loginAs, noHorizontalScroll, open, root, settle, toasts } from '../../lib/console';
import { restoreConfig, sql } from '../../lib/db';
import { API_URL, CONSOLE_URL } from '../../lib/env';
import { stargameBuilt } from '../../lib/samples';

const HTTP_KEY = process.env.HTTP_KEY || 'defaulthttpkey';
const PATH = '/sample-data/stargame';
const KEY_COL: Record<string, string> = { economy_gift: 'code', base_app_version: 'platform' };
const k = uid().replace(/_/g, '');

test.describe.configure({ mode: 'serial' });

let switchedOff: { table: string; key: string }[] = [];
let referralHeld = false;

async function serverRpc(id: string, payload: unknown) {
  const r = await fetch(API_URL + '/v2/rpc/' + id + '?unwrap&http_key=' + encodeURIComponent(HTTP_KEY), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
  const text = await r.text();
  let data: any = null;
  try { data = JSON.parse(text); } catch { data = text; }
  return { status: r.status, data };
}

function releaseReferral() {
  if (referralHeld) sql(`UPDATE economy_referral_settings SET goal_value=goal_value-1000`);
  referralHeld = false;
}

function restoreSwitchedOff() {
  for (const r of switchedOff) {
    const col = KEY_COL[r.table] || 'key';
    sql(`UPDATE ${r.table} SET is_active=true WHERE ${col}='${r.key.replace(/'/g, "''")}'`);
  }
  switchedOff = [];
}

test.beforeAll(() => {
  test.skip(!stargameBuilt(), 'stargame is excluded in this fork');
});

test.afterAll(async () => {
  if (!stargameBuilt()) return;
  const seeded = await serverRpc('stargame_seed', {});
  expect(seeded.status).toBe(200);
  restoreSwitchedOff();
  releaseReferral();
  restoreConfig();
});

test('the preview lists the sample rows by table and changes nothing', async () => {
  const before = sql(`SELECT count(*) FROM economy_product WHERE key LIKE 'star\\_%'`);
  const d = await base.get(PATH);
  expect(d.sample).toBe('stargame');
  expect(d.admin).toBe(true);
  expect(d.pluginLoaded).toBe(true);
  expect(d.preview.applied).toBe(false);
  expect(d.preview.deleted).toBeGreaterThan(50);
  const banners = d.preview.tables.find((x: any) => x.kind === 'banner');
  expect(banners.deleted).toEqual(expect.arrayContaining(['star_bn_welcome', 'star_bn_nova']));
  const gifts = d.preview.tables.find((x: any) => x.kind === 'gift');
  expect([...gifts.deleted, ...gifts.kept.map((x: any) => x.key)]).toEqual(expect.arrayContaining(['STAR100', 'HALF50', 'STARFAN']));
  for (const tb of d.preview.tables) {
    expect(['language', 'currency', 'meta']).not.toContain(tb.kind);
    expect(tb.deleted).not.toContain('coin');
  }
  expect(sql(`SELECT count(*) FROM economy_product WHERE key LIKE 'star\\_%'`)).toBe(before);
});

test('the page is in Ctrl K, shows the preview and needs an admin to run', async ({ page, spawn }) => {
  await open(page, '#/system/updates');
  await page.keyboard.press('Control+k');
  await page.locator('.pl-pal-input').fill('داده‌های نمونه');
  await expect(page.locator('.pl-pal-row').first()).toContainText('داده‌های نمونه');
  await page.keyboard.press('Enter');
  await expect(page).toHaveURL(/#\/system\/sample-data/);
  await settle(page);
  const box = root(page).locator('[data-sample-page="stargame"]');
  await expect(box).toBeVisible();
  await expect(box.locator('h2')).toHaveText('حذف داده‌های نمونهٔ ستاره‌چین');
  await expect(box.locator('[data-sc-plugin="on"]')).toContainText('STARGAME_SEED=false');
  await expect(box.locator('[data-sc-stat="deleted"]')).not.toContainText(/^۰/);
  const row = box.locator('tr', { hasText: 'publication_banner' }).first();
  await expect(row).toContainText('بنرها');
  await expect(row).toContainText('star_bn_welcome');
  await expect(box.locator('[data-sc-run]')).toBeEnabled();

  const who = { user: 'e2esc' + k.slice(-8), pass: 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!' };
  const add = await consoleApi('POST', '/v2/console/user', { username: who.user, email: who.user + '@example.com', acl: { ACCOUNT: { read: true }, SETTINGS: { read: true, write: true } } });
  expect(add.status).toBeLessThan(300);
  try {
    const reset = await must(consoleApi('POST', '/v2/console/user/' + who.user + '/reset/password'));
    const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: reset.code || add.data?.token, password: who.pass }) });
    expect(pc.ok).toBe(true);
    const auth = await (await fetch(CONSOLE_URL + '/v2/console/authenticate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username: who.user, password: who.pass }) })).json();
    const bearer = 'Bearer ' + auth.token;
    const view = await consoleApi('GET', '/v2/console/base' + PATH, undefined, bearer);
    expect(view.status).toBe(200);
    expect(view.data.admin).toBe(false);
    const refused = await consoleApi('POST', '/v2/console/base' + PATH + '/cleanup', { confirm: 'stargame' }, bearer);
    expect(refused.status).toBe(403);
    const p = await spawn({ signedIn: false });
    await loginAs(p, who);
    await p.goto(CONSOLE_URL + '/#/system/sample-data');
    await settle(p);
    await expect(root(p).locator('[data-sc-admin-only]')).toContainText('فقط مدیر کنسول');
    await expect(root(p).locator('[data-sc-run]')).toBeDisabled();
  } finally {
    await consoleApi('DELETE', '/v2/console/user?username=' + who.user);
  }
  expect(Number(sql(`SELECT count(*) FROM economy_gift WHERE code='STAR100'`))).toBe(1);
});

test('the server RPC refuses players and previews for a server call', async () => {
  const player = await devicePlayer('e2e-sc-' + k);
  const r = await rpc(player, 'base_sample_cleanup', { apply: true });
  await deleteUser(player.id);
  expect(r.status).toBeGreaterThanOrEqual(400);
  expect(JSON.stringify(r.data)).toContain('SERVER_ONLY');
  const wrong = await serverRpc('base_sample_cleanup', { sample: 'other' });
  expect(JSON.stringify(wrong.data)).toContain('SAMPLE_UNKNOWN');
  const preview = await serverRpc('base_sample_cleanup', {});
  expect(preview.status).toBe(200);
  expect(preview.data.result.applied).toBe(false);
  expect(preview.data.result.deleted).toBeGreaterThan(50);
  expect(Number(sql(`SELECT count(*) FROM economy_gift WHERE code='STAR100'`))).toBe(1);
});

test('an admin removes the sample data from the page, and a second run is a no-op', async ({ page }) => {
  sql(`UPDATE economy_referral_settings SET goal_value=goal_value+1000`);
  referralHeld = true;
  const d = await base.get(PATH);
  expect(d.preview.changes.find((c: any) => c.kind === 'referral_rule')).toBeUndefined();
  for (const tb of d.preview.tables) for (const kp of tb.kept) if (kp.reason === 'in_use_switched_off') switchedOff.push({ table: tb.table, key: kp.key });
  const deleted = d.preview.deleted;
  await open(page, '#/system/sample-data');
  const box = root(page).locator('[data-sample-page="stargame"]');
  await box.locator('[data-sc-run]').click();
  const dialog = page.locator('[role="alertdialog"]');
  await expect(dialog).toContainText('داده‌های نمونه حذف شود؟');
  await expect(dialog).toContainText('یک تراکنش');
  await dialog.getByRole('button', { name: 'حذف کن' }).click();
  await expect(toasts(page).last()).toContainText('داده‌های نمونه حذف شد');
  await settle(page);
  await expect(box.locator('[data-sc-last]')).toHaveAttribute('data-sc-last', /^console:/);
  await expect(box.locator('[data-sc-stat="deleted"]')).toContainText('۰');
  await expect(box.locator('[data-sc-run]')).toBeDisabled();

  for (const [q, want] of [
    [`SELECT count(*) FROM publication_banner WHERE key LIKE 'star\\_bn\\_%'`, '0'],
    [`SELECT count(*) FROM base_config_key WHERE key LIKE 'stargame.%'`, '0'],
    [`SELECT count(*) FROM economy_chance_game WHERE key IN ('star_wheel_daily','star_lucky_box')`, '0'],
    [`SELECT count(*) FROM base_tag WHERE key='star-vip'`, '0'],
    [`SELECT count(*) FROM base_language WHERE code IN ('fa','en','ar')`, '3'],
    [`SELECT count(*) FROM economy_currency WHERE key='coin'`, '1'],
    [`SELECT count(*) FROM leaderboard WHERE id='stargame_tictactoe_wins'`, '0'],
  ] as const) expect(sql(q), q).toBe(want);

  const last = await base.get(PATH);
  expect(last.lastRun.by).toMatch(/^console:/);
  expect(last.lastRun.deleted).toBe(deleted);
  expect(last.preview.deleted).toBe(0);
  expect(last.preview.changed).toBe(0);

  const again = await must(consoleApi('POST', '/v2/console/base' + PATH + '/cleanup', { confirm: 'stargame' }));
  expect(again.result.deleted).toBe(0);
  expect(again.result.changed).toBe(0);
  const noConfirm = await consoleApi('POST', '/v2/console/base' + PATH + '/cleanup', {});
  expect(noConfirm.status).toBe(400);

  const rpcRun = await serverRpc('base_sample_cleanup', { apply: true, by: 'deploy' });
  expect(rpcRun.status).toBe(200);
  expect(rpcRun.data.result.deleted).toBe(0);
  expect((await base.get(PATH)).lastRun.by).toBe('rpc:deploy');

});

test('English and dark mobile layout', async ({ spawn }) => {
  const p = await spawn({ signedIn: true, locale: 'en', mode: 'dark', mobile: true, width: 390, height: 844 });
  await open(p, '#/system/sample-data');
  const box = root(p).locator('[data-sample-page="stargame"]');
  await expect(box.locator('h2')).toHaveText('Remove the Star Catcher sample data');
  await expect(box.locator('[data-sc-stat="deleted"]')).toContainText('0');
  await noHorizontalScroll(p);
});

test('the sample comes back with the stargame seed RPC', async () => {
  const seeded = await serverRpc('stargame_seed', {});
  expect(seeded.status).toBe(200);
  expect(seeded.data.created).toEqual(expect.arrayContaining(['gift:STAR100', 'banner:star_bn_welcome']));
  restoreSwitchedOff();
  releaseReferral();
  expect(sql(`SELECT count(*) FROM leaderboard WHERE id='stargame_tictactoe_wins'`)).toBe('1');
  const d = await base.get(PATH);
  expect(d.preview.deleted).toBeGreaterThan(50);
});
