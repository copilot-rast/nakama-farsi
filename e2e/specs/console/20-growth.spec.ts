import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { open, openSettings, psWindow, root, rows, settle, toasts } from '../../lib/console';
import { restoreConfig } from '../../lib/db';

const k = uid();
const players: string[] = [];

test.describe.configure({ mode: 'serial' });

test.afterAll(async () => {
  for (const id of players) await deleteUser(id);
  restoreConfig();
});

async function newPlayer(tag: string): Promise<Player> {
  const p = await devicePlayer(`e2e-gr-${tag}-${k}`);
  players.push(p.id);
  return p;
}

async function configValue(p: Player, key: string, platform: string, version = '1.4.0') {
  const r = await rpc(p, 'base_remote_config', { platform, version, keys: [key] });
  expect(r.status, JSON.stringify(r.data)).toBe(200);
  return { value: r.data.values[key], variant: r.data.variants[key] };
}

test.describe('growth: remote config', () => {
  test('a key with a web rule is created, answered by the client RPC, edited, rolled back and previewed', async ({ page }) => {
    const key = k.replace(/[^a-z0-9_]/g, '') + '.sale';
    const p = await newPlayer('cfg');
    await open(page, '#/publication/remote-config');
    await expect(rows(page).filter({ hasText: 'stargame.banner' })).toContainText('متن ترجمه‌شده');
    await root(page).locator('[data-config-new]').click();
    await expect(page).toHaveURL(/#\/publication\/remote-config\/new/);
    const ed = root(page).locator('[data-config-editor]');
    await ed.locator('[data-config-key]').fill('Bad Key');
    await ed.locator('[data-config-save]').click();
    await expect(ed).toContainText('این کلید معتبر نیست');
    await ed.locator('[data-config-key]').fill(key);
    await ed.locator('[data-config-type]').selectOption('bool');
    await ed.locator('[data-rule-add]').click();
    const rule = ed.locator('[data-rule]').first();
    await rule.locator('input[aria-label="نام قانون"]').fill('web-new');
    await rule.locator('label', { hasText: 'وب' }).click();
    await rule.locator('input[aria-label="از نسخهٔ"]').fill('1.4');
    await ed.locator('[data-config-save]').click();
    await expect(page).toHaveURL(/#\/publication\/remote-config\/[0-9a-f-]{36}/);
    await expect(toasts(page).last()).toContainText('ذخیره');
    expect(await configValue(p, key, 'web')).toEqual({ value: true, variant: 'web-new' });
    expect(await configValue(p, key, 'web', '1.3.9')).toEqual({ value: false, variant: 'default' });
    expect(await configValue(p, key, 'android')).toEqual({ value: false, variant: 'default' });

    await settle(page);
    const ed2 = root(page).locator('[data-config-editor]');
    await ed2.locator('[data-rule] input[aria-label="درصد بازیکنان"]').fill('۰');
    await ed2.locator('[data-config-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    expect(await configValue(p, key, 'web')).toEqual({ value: false, variant: 'default' });
    const hist = ed2.locator('[data-config-history]');
    await expect(hist.locator('[data-history-version]')).toHaveCount(2);
    await hist.locator('[data-rollback="1"]').click();
    await page.locator('[data-economy="dialog"]').getByRole('button', { name: 'برگشت به این نسخه' }).click();
    await expect(toasts(page).last()).toContainText('به نسخهٔ ۱ برگشت');
    expect(await configValue(p, key, 'web')).toEqual({ value: true, variant: 'web-new' });
    await expect(root(page).locator('[data-config-history] [data-history-version]')).toHaveCount(3);

    const prev = root(page).locator('[data-config-preview]');
    await prev.locator('input[aria-label="شناسهٔ بازیکن"]').fill(p.id);
    await prev.locator('input[aria-label="نسخهٔ برنامه"]').fill('1.5.0');
    await prev.locator('[data-preview-run]').click();
    await expect(prev.locator(`[data-preview-key="${key}"]`)).toContainText('web-new');
    await expect(prev.locator('[data-preview-key="stargame.golden_rush"]')).toContainText('true');

    await open(page, '#/publication/remote-config');
    await expect(root(page).locator('[data-config-recent]')).toContainText(key);
    const id = (await base.get('/remote-config?q=' + key)).items[0].id;
    await base.del('/remote-config/' + id);
    expect((await rpc(p, 'base_remote_config', { platform: 'web', keys: [key] })).data.values[key]).toBeUndefined();
  });
});

test.describe('growth: player activity', () => {
  test('sign-ins and tickets show in the activity log, filter by kind, and on the player page card', async ({ page }) => {
    const p = await newPlayer('act');
    await rpc(p, 'base_support_ticket_create', { subject: 'فعالیت ' + k, body: 'سلام' });
    await open(page, '#/moderation/activity?userId=' + p.id);
    await expect(rows(page).filter({ hasText: 'ورود' }).first()).toContainText('با دستگاه وارد شد');
    await expect(rows(page).filter({ hasText: 'فعالیت ' + k })).toHaveCount(1);
    await root(page).locator('.eco-filters select').first().selectOption('ticket');
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first()).toContainText('خود بازیکن');
    await page.goto(page.url().split('#')[0] + '#/players/' + p.id);
    const card = page.locator('[data-player-activity]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('[data-activity-item="ticket"]')).toContainText('فعالیت ' + k);
    await card.locator('[data-activity-chips] [data-value="signin"]').click();
    await expect(card.locator('[data-activity-item="ticket"]')).toHaveCount(0);
    await expect(card.locator('[data-activity-item="signin"]').first()).toBeVisible();
    await card.locator('[data-activity-all]').click();
    await expect(page).toHaveURL(new RegExp('#/moderation/activity\\?userId=' + p.id));
  });
});

test.describe('growth: referrals', () => {
  test('a referral waits for the goal, pays both players once, and a held referral is approved or rejected in the console', async ({ page }) => {
    const initial = await eco.get('/referral-settings');
    await eco.put('/referral-settings', { ...initial, holdFlagged: false });
    const a = await newPlayer('ra');
    const b = await newPlayer('rb');
    const code = (await rpc(a, 'economy_referral_me', {})).data.code as string;
    expect(code).toMatch(/^[A-Z2-9]{8}$/);
    const entered = await rpc(b, 'economy_referral_enter', { code: code.toLowerCase() });
    expect(entered.data.entered.status).toBe('pending');
    await open(page, '#/economy/referrals?referrerId=' + a.id);
    await expect(root(page).locator('[data-referral-top]')).toContainText('همهٔ دعوت‌ها');
    const row = rows(page).filter({ hasText: b.username });
    await expect(row.locator('[data-referral-status]')).toHaveText('منتظر شرط');
    await expect(row).toContainText('round_finished');
    await expect(row).toContainText('۰ از ۱');
    const walletA = (await eco.get('/users/' + a.id)).wallet?.coin ?? 0;
    await eco.post('/users/' + b.id + '/events', { key: 'round_finished', value: 10 });
    await page.reload();
    await settle(page);
    await expect(rows(page).filter({ hasText: b.username }).locator('[data-referral-status]')).toHaveText('کامل');
    expect(((await eco.get('/users/' + a.id)).wallet?.coin ?? 0) - walletA).toBe(100);
    await eco.post('/users/' + b.id + '/events', { key: 'round_finished', value: 10 });
    expect(((await eco.get('/users/' + a.id)).wallet?.coin ?? 0) - walletA).toBe(100);

    const settings = await eco.get('/referral-settings');
    await eco.put('/referral-settings', { ...settings, dailyPerReferrer: 1, holdFlagged: true });
    const c = await newPlayer('rc');
    const d = await newPlayer('rd');
    expect((await rpc(c, 'economy_referral_enter', { code })).data.entered.status).toBe('held');
    expect((await rpc(d, 'economy_referral_enter', { code })).data.entered.status).toBe('held');
    await open(page, '#/economy/referrals?status=held&referrerId=' + a.id);
    const held = rows(page).filter({ hasText: c.username });
    await expect(held.locator('[data-flag="burst"]')).toHaveText('دعوت زیاد در یک روز');
    await held.locator('[data-approve]').click();
    await page.locator('[data-form-dialog="referral-approve"] [data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('دعوت تأیید شد');
    const other = rows(page).filter({ hasText: d.username });
    await other.locator('[data-reject]').click();
    await page.locator('[data-form-dialog="referral-reject"] textarea').fill('همان آدم');
    await page.locator('[data-form-dialog="referral-reject"] [data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('دعوت رد شد');
    const list = await eco.get('/referrals?referrerId=' + a.id);
    const byName = Object.fromEntries(list.items.map((x: any) => [x.refereeName, x.status]));
    expect(byName).toMatchObject({ [b.username]: 'completed', [c.username]: 'pending', [d.username]: 'rejected' });

    await open(page, '#/economy/referrals?referrerId=' + a.id);
    await rows(page).first().locator('[data-tree]').click();
    const tree = page.locator('[data-referral-tree="' + a.id + '"]');
    await expect(tree.locator('[data-tree-child]')).toHaveCount(3);
    await page.keyboard.press('Escape');

    await openSettings(page, 'referral-rules');
    const form = psWindow(page).locator('[data-settings-form="referral-rules"]');
    await expect(form).toContainText('referral_inviter');
    await expect(form).toContainText('round_finished');
    restoreConfig();
  });
});
