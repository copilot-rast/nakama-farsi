import { expect, test } from '../../lib/fixtures';
import type { Locator } from '@playwright/test';
import { deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, open, panel, savePanel, settlePanel } from '../../lib/console';

const k = uid();
const metaKey = 'e2e_' + k + '_xp';
const boostKey = 'e2e_' + k + '_boost';
const streakKey = 'e2e_' + k + '_track';
const eventKey = 'e2e_' + k + '_match';
const made = { users: [] as string[], prizes: [] as string[], metas: [] as string[], rules: [] as string[], types: [] as string[], streaks: [] as string[] };
let P: Player;

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

async function pickIn(scope: Locator, text: string, exact?: RegExp) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = exact ? pop.locator('[role="option"]').filter({ has: scope.page().locator('span.font-mono', { hasText: exact }) }).first() : pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

function xp() {
  return Number(sql(`SELECT COALESCE((SELECT value::int FROM economy_user_meta um JOIN economy_meta m ON m.id=um.meta_id WHERE um.user_id='${P.id}' AND m.key='${metaKey}'), 0)`));
}

test.beforeAll(async () => {
  P = await devicePlayer('e2e-bs-' + k, ('bs' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  const meta = await eco.post('/metas', { key: metaKey, name: { fa: 'تجربهٔ آزمون ' + k, en: 'Test XP ' + k } });
  made.metas.push(meta.id);
  const type = await eco.post('/event-types', { key: eventKey, name: { fa: 'پایان بازی ' + k }, clientAllowed: false, isActive: true, rateLimit: 1000 });
  made.types.push(type.id);
  const rule = await eco.post('/meta-rules', { key: 'e2e_' + k + '_rule', name: { fa: 'تجربه از بازی' }, eventTypeId: type.id, metaId: meta.id, action: 'add', valueSource: 'fixed', amount: 10, isActive: true });
  made.rules.push(rule.id);
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_reward_claim WHERE owner_id IN (SELECT id FROM economy_streak WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_user_prize WHERE prize_id IN (SELECT id FROM economy_prize WHERE key LIKE 'e2e\\_${k}%')`);
  for (const it of (await eco.get('/streaks?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/streaks/' + it.id)).status).toBe(200);
  for (const id of made.rules) expect((await eco.del('/meta-rules/' + id)).status).toBe(200);
  for (const it of (await eco.get('/prizes?q=e2e_' + k + '&limit=50')).items || []) expect((await eco.del('/prizes/' + it.id)).status).toBe(200);
  sql(`DELETE FROM economy_player_event WHERE type_id IN (SELECT id FROM economy_event_type WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_player_event_total WHERE type_id IN (SELECT id FROM economy_event_type WHERE key LIKE 'e2e\\_${k}%')`);
  sql(`DELETE FROM economy_player_event_day WHERE type_id IN (SELECT id FROM economy_event_type WHERE key LIKE 'e2e\\_${k}%')`);
  for (const id of made.types) await eco.del('/event-types/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
});

test.describe('boosters and streak extras', () => {
  test('console: a prize gives a timed booster; the booster doubles meta rule XP and does not stack', async ({ page }) => {
    await open(page, '#/economy/prizes/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(boostKey);
    await fillTranslated(p, 'name', { fa: 'دوبرابر تجربه ' + k });
    const row = p.locator('.eco-items-row').first();
    await row.locator('select').first().selectOption('booster');
    await pickIn(row, metaKey, new RegExp('^' + metaKey + '$'));
    const val = row.locator('[data-booster-value]');
    await expect(val).toContainText('ضریب');
    await val.locator('input').nth(0).fill('2');
    await val.locator('input').nth(1).fill('60');
    await savePanel(page);
    const prize = await byKey('prizes', boostKey);
    expect(prize.items).toEqual([{ type: 'booster', key: metaKey, action: 'give', value: { multiplier: 2, minutes: 60 } }]);
    made.prizes.push(prize.id);
    await eco.put('/prizes/' + prize.id, Object.assign({}, prize, { isRepeatable: true }));

    await eco.post('/users/' + P.id + '/events', { key: eventKey, value: 1 });
    expect(xp()).toBe(10);
    await eco.post('/users/' + P.id + '/grant', { prizeId: prize.id });
    await eco.post('/users/' + P.id + '/events', { key: eventKey, value: 1 });
    expect(xp()).toBe(30);
    await eco.post('/users/' + P.id + '/grant', { prizeId: prize.id });
    const list = (await rpc(P, 'economy_boosters', { lang: 'fa' })).data.items;
    expect(list).toHaveLength(1);
    expect(list[0]).toMatchObject({ metaKey, multiplier: 2, metaName: 'تجربهٔ آزمون ' + k });
    expect(list[0].secondsLeft).toBeGreaterThan(119 * 60);
    await eco.post('/users/' + P.id + '/events', { key: eventKey, value: 1 });
    expect(xp()).toBe(50);
  });

  test('console: streak with week start, special days and an infinity phase; the player view shows them', async ({ page }) => {
    const day = (await eco.post('/prizes', { key: 'e2e_' + k + '_day', name: { fa: 'روز ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }] })).id;
    const extra = (await eco.post('/prizes', { key: 'e2e_' + k + '_extra', name: { fa: 'ویژه ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 100 }] })).id;
    await open(page, '#/economy/streaks');
    await go(page, '#/economy/streaks/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(streakKey);
    await fillTranslated(p, 'name', { fa: 'جایزهٔ روزانهٔ آزمون ' + k });
    await pickIn(field(p, 'رویداد شمارنده'), eventKey);
    await field(p, 'اگر بیشتر غیبت کند').locator('select').selectOption('week_start');
    await expect(field(p, 'طول هفته (روز)')).toBeVisible();
    await field(p, 'طول هفته (روز)').locator('input').fill('3');
    const rewards = p.locator('[data-rows="rewards"]');
    await rewards.locator('[data-rows-add="rewards"]').click();
    await pickIn(rewards.locator('[data-row="0"] [data-col="prizeId"]'), 'روز ' + k);
    const extras = p.locator('[data-rows="extras"]');
    await expect(p.locator('[data-extras-note]')).toContainText('بدون غیبت');
    await extras.locator('[data-rows-add="extras"]').click();
    await extras.locator('[data-row="0"] [data-col="day"] input').fill('2');
    await pickIn(extras.locator('[data-row="0"] [data-col="prizeId"]'), 'ویژه ' + k);
    await p.locator('label.x-toggle', { hasText: 'بعد از پایان چرخه، مرحلهٔ بی‌پایان' }).click();
    await pickIn(field(p, /^ارز$/), 'coin', /^coin$/);
    await savePanel(page);
    const sk = await byKey('streaks', streakKey);
    expect(sk).toMatchObject({ missAction: 'week_start', weekLength: 3, infinity: true, infinityCurrency: 'coin', infinityMin: 150, infinityMax: 180, cycleLength: 7 });
    expect(sk.extras).toEqual([expect.objectContaining({ day: 2, prizeId: extra, grantMode: 'claim' })]);
    expect(sk.rewards[0]).toMatchObject({ day: 1, prizeId: day });

    await eco.post('/users/' + P.id + '/events', { key: eventKey, value: 1 });
    const view = (await rpc(P, 'economy_streaks', { lang: 'fa' })).data.items.find((x: any) => x.key === streakKey);
    expect(view).toMatchObject({ count: 1, missAction: 'week_start', weekLength: 3, phase: 'track', infinity: { currencyKey: 'coin', min: 150, max: 180 } });
    expect(view.extras).toEqual([expect.objectContaining({ day: 2, status: 'locked', grantMode: 'claim' })]);
  });
});
