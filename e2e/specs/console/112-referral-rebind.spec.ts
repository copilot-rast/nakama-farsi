import { expect, test } from '../../lib/fixtures';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { open, openSettings, psWindow, rows } from '../../lib/console';
import { sql } from '../../lib/db';

const k = uid();
const players: string[] = [];
let initial: Record<string, unknown> = {};
const ids: string[] = [];

test.describe.configure({ mode: 'serial' });

test.beforeAll(async () => {
  initial = await eco.get('/referral-settings');
});

test.afterAll(async () => {
  await eco.put('/referral-settings', initial);
  for (const id of players) await deleteUser(id);
  for (const id of ids) sql(`DELETE FROM economy_event_type WHERE id='${id}'`);
});

async function newPlayer(tag: string, display: string): Promise<Player> {
  const p = await devicePlayer(`e2e-rb-${tag}-${k}`);
  players.push(p.id);
  sql(`UPDATE users SET display_name='${display}' WHERE id='${p.id}'`);
  return p;
}

async function eventType(key: string): Promise<string> {
  const r = await eco.post('/event-types', { key, name: { fa: key, en: key }, clientAllowed: true, isActive: true, rateLimit: 1000 });
  ids.push(r.id);
  return r.id;
}

test('referral list shows display names and old invites move to the current rule', async ({ page }) => {
  const oldGoal = await eventType(`rb_old_${k}`.replace(/[^a-z0-9_]/g, ''));
  const newGoal = await eventType(`rb_new_${k}`.replace(/[^a-z0-9_]/g, ''));
  await eco.put('/referral-settings', { ...initial, isEnabled: true, holdFlagged: false, onePerDevice: false, onePerPhone: false, goalEventTypeId: oldGoal, goalMeasure: 'count', goalValue: 1 });
  const a = await newPlayer('a', 'دعوت‌کننده');
  const b = await newPlayer('b', 'راننده تازه');
  const code = (await rpc(a, 'economy_referral_me', {})).data.code as string;
  expect((await rpc(b, 'economy_referral_enter', { code })).data.entered.status).toBe('pending');
  const me = await rpc(a, 'economy_referral_me', {});
  expect(me.data.recent[0].displayName).toBe('راننده تازه');

  await open(page, '#/economy/referrals?referrerId=' + a.id);
  const row = rows(page).filter({ hasText: b.username });
  await expect(row.locator(`[data-referral-person="${b.id}"] [data-display-name]`)).toHaveText('راننده تازه');
  await expect(row.locator(`[data-referral-person="${a.id}"] [data-display-name]`)).toHaveText('دعوت‌کننده');

  await eco.put('/referral-settings', { ...(await eco.get('/referral-settings')), goalEventTypeId: newGoal });
  await rpc(b, 'economy_event_record', { key: `rb_new_${k}`.replace(/[^a-z0-9_]/g, ''), value: 1 });
  expect((await eco.get('/referrals?referrerId=' + a.id)).items[0].status).toBe('pending');

  await openSettings(page, 'referral-rules');
  const card = psWindow(page).locator('[data-referral-rebind]');
  await expect(card.locator('[data-rebind-total]')).toBeVisible();
  await expect(card).toContainText('دعوت‌های قدیمی با شرط قبلی');
  await card.locator('[data-rebind-apply]').click();
  await page.locator('[data-form-dialog="referral-rebind"] [data-dialog-ok]').click();
  await expect(card.locator('[data-rebind-none]')).toBeVisible();
  expect((await eco.get('/referrals?referrerId=' + a.id)).items[0].status).toBe('completed');
});

