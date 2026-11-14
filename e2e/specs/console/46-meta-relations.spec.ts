import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, devicePlayer, eco, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, open, openSettings, panel, psWindow, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { users: [] as string[], prizes: [] as string[], types: [] as string[], rels: [] as string[], metaRules: [] as string[], metas: [] as string[] };
const M: Record<string, any> = {};
let prize: any;
let P: Player;
let settings: any;

test.describe.configure({ mode: 'serial' });

async function byKey(resource: string, key: string) {
  const res = await eco.get('/' + resource + '?q=' + encodeURIComponent(key) + '&limit=50');
  return (res.items || []).find((x: any) => x.key === key);
}

async function pickIn(scope: Locator, text: string) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

function studio(page: Page) {
  return page.locator('[data-meta-studio]');
}

async function settleStudio(page: Page) {
  await expect(studio(page)).toBeVisible();
  await expect(panel(page).locator('.animate-pulse')).toHaveCount(0);
}

test.beforeAll(async () => {
  for (const [name, def] of [['xp', 0], ['lvl', 1], ['gold', 0]] as const) {
    M[name] = await eco.post('/metas', { key: 'e2e_' + k + '_' + name, name: { fa: name + ' ' + k.slice(-4), en: name + ' ' + k.slice(-4) }, defaultValue: def, serverOnly: false, settings: {} });
    made.metas.push(M[name].id);
  }
  prize = await eco.post('/prizes', { key: 'e2e_' + k + '_rp', name: { fa: 'جایزهٔ رابطه ' + k, en: 'Relation prize ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 5 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  const et = await eco.post('/event-types', { key: 'e2e_' + k + '_ev', name: { fa: 'رویداد ' + k, en: 'Event ' + k }, clientAllowed: false, isActive: true, rateLimit: 100, sort: 0 });
  made.types.push(et.id);
  settings = await eco.get('/event-settings');
  P = await devicePlayer('e2e-rel-' + k, ('rel' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  await eco.put('/event-settings', settings).catch(() => null);
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_reward_claim WHERE kind='relation' AND prize_id='${prize.id}'; DELETE FROM economy_user_prize WHERE prize_id='${prize.id}'`);
  for (const id of made.rels) await eco.del('/meta-relations/' + id);
  for (const id of made.metaRules) await eco.del('/meta-rules/' + id);
  for (const id of made.types) await eco.del('/event-types/' + id);
  for (const id of made.metas) await eco.del('/metas/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
});

test.describe('meta relations: tab and editor', () => {
  test('the metas hub has a relations tab, the old address redirects, and a relation is built in its editor', async ({ page }) => {
    await open(page, '#/economy/metas');
    const tab = root(page).locator('[data-economy-tab="meta-relations"]');
    await expect(tab).toHaveText('رابطهٔ متاها');
    await tab.click();
    await expect(page).toHaveURL(/#\/economy\/metas\/relations/);
    await page.goto(CONSOLE_URL + '/#/economy/meta-relations');
    await expect(page).toHaveURL(/#\/economy\/metas\/relations/);
    await root(page).getByRole('button', { name: 'رابطهٔ تازه' }).first().click();
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill('e2e_' + k + '_up');
    await fillTranslated(p, 'name', { fa: 'بالا رفتن ' + k, en: 'Level up ' + k });
    await pickIn(field(p, 'متای شرط'), 'e2e_' + k + '_xp');
    await field(p, 'مقایسه').locator('select').selectOption('gte');
    await field(p, /^عدد$/).locator('input').fill('۱۰۰');
    const targets = p.locator('[data-rows="targets"]');
    await expect(targets.locator('[data-row]')).toHaveCount(1);
    await pickIn(targets.locator('[data-row="0"] [data-col="metaId"]'), 'e2e_' + k + '_lvl');
    await targets.locator('[data-rows-add="targets"]').click();
    await pickIn(targets.locator('[data-row="1"] [data-col="metaId"]'), 'e2e_' + k + '_xp');
    await targets.locator('[data-row="1"] [data-col="amount"] input').fill('-100');
    await field(p, 'نوع جایزه').locator('select').selectOption('prize');
    await pickIn(field(p, /^جایزه$/), 'e2e_' + k + '_rp');
    await field(p, 'روش دادن').locator('select').selectOption('claim');
    await field(p, 'چند بار').locator('select').selectOption('rearm');
    await field(p, 'سقف دفعه‌ها').locator('input').fill('۳');
    const preview = p.locator('[data-rel-preview]');
    await expect(preview).toContainText('≥ ۱۰۰');
    await expect(preview).toContainText('هر بار که شرط از نو برقرار شود');
    await savePanel(page);
    const rel = await byKey('meta-relations', 'e2e_' + k + '_up');
    made.rels.push(rel.id);
    expect(rel).toMatchObject({ metaId: M.xp.id, operator: 'gte', value: 100, prizeType: 'prize', prizeId: prize.id, grantMode: 'claim', repeat: 'rearm', maxFires: 3, isActive: true, clientVisible: true });
    expect(rel.targets).toEqual([{ metaId: M.lvl.id, action: 'add', amount: 1 }, { metaId: M.xp.id, action: 'add', amount: -100 }]);
    await search(page, 'e2e_' + k + '_up');
    const row = rows(page).first();
    await expect(row).toContainText('بالا رفتن ' + k);
    await expect(row.locator('[data-rel-when]')).toContainText('≥ ۱۰۰');
    await expect(row.locator('[data-rel-target]')).toHaveCount(2);
    await expect(row).toContainText('دوباره پس از افت');
  });

  test('the editor asks for a result and refuses a duplicate target', async ({ page }) => {
    await open(page, '#/economy/metas/relations/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, /^کلید$/).locator('input').fill('e2e_' + k + '_bad');
    await fillTranslated(p, 'name', { fa: 'نادرست' });
    await pickIn(field(p, 'متای شرط'), 'e2e_' + k + '_gold');
    const targets = p.locator('[data-rows="targets"]');
    await targets.locator('[data-row="0"] button[title="حذف"]').click();
    await expect(p.locator('[data-rel-needs]')).toContainText('دست‌کم یک متای نتیجه یا یک جایزه');
    await targets.locator('[data-rows-add="targets"]').click();
    await targets.locator('[data-rows-add="targets"]').click();
    await pickIn(targets.locator('[data-row="0"] [data-col="metaId"]'), 'e2e_' + k + '_lvl');
    await pickIn(targets.locator('[data-row="1"] [data-col="metaId"]'), 'e2e_' + k + '_lvl');
    await expect(p.locator('[data-rel-needs]')).toHaveCount(0);
    await savePanel(page).catch(() => null);
    await expect(targets.locator('[data-row="1"] .eco-items-err')).toContainText('هر متا با هر نوع تغییر فقط یک بار');
    expect(await byKey('meta-relations', 'e2e_' + k + '_bad')).toBeUndefined();
  });
});

test.describe('meta relations: the meta studio', () => {
  test('a meta row opens its studio with inputs and outputs; items are added and edited in place with the meta locked', async ({ page }) => {
    await open(page, '#/economy/metas');
    await search(page, 'e2e_' + k + '_xp');
    await rows(page).first().locator('td').first().click();
    await settleStudio(page);
    await expect(page).toHaveURL(new RegExp('studio=' + M.xp.id));
    await expect(panel(page)).toContainText('استودیوی xp ' + k.slice(-4));
    const out = studio(page).locator('[data-studio-group="relations-out"]');
    await expect(out).toContainText('بالا رفتن ' + k);
    await expect(studio(page).locator('[data-studio-group="meta-rules"]')).toContainText('هیچ رویدادی این متا را تغییر نمی‌دهد');
    await expect(studio(page).locator('[data-studio-node]')).toContainText('e2e_' + k + '_xp');

    await studio(page).locator('[data-studio-add="meta-rules"]').click();
    const ed = studio(page).locator('[data-studio-editor="meta-rules"]');
    await expect(ed).toBeVisible();
    await expect(ed.locator('[data-locked="metaId"]')).toContainText('e2e_' + k + '_xp');
    await field(ed, /^کلید$/).locator('input').fill('e2e_' + k + '_mr');
    await fillTranslated(ed, 'name', { fa: 'تجربه از رویداد ' + k });
    await pickIn(field(ed, 'رویداد'), 'e2e_' + k + '_ev');
    await field(ed, /^عدد$/).locator('input').fill('۳۰');
    await panel(page).locator('[data-studio-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    await settleStudio(page);
    const rules = studio(page).locator('[data-studio-group="meta-rules"]');
    await expect(rules).toContainText('تجربه از رویداد ' + k);
    const mr = await byKey('meta-rules', 'e2e_' + k + '_mr');
    made.metaRules.push(mr.id);
    expect(mr).toMatchObject({ metaId: M.xp.id, amount: 30 });

    await out.locator('[data-studio-item]').first().click();
    const red = studio(page).locator('[data-studio-editor="meta-relations"]');
    await expect(red).toBeVisible();
    await expect(red.locator('[data-locked="metaId"]')).toContainText('e2e_' + k + '_xp');
    await field(red, /^عدد$/).locator('input').fill('۲۰۰');
    await panel(page).locator('[data-studio-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    await settleStudio(page);
    expect((await byKey('meta-relations', 'e2e_' + k + '_up')).value).toBe(200);
    await expect(out.locator('[data-rel-when]')).toContainText('≥ ۲۰۰');

    await studio(page).locator('[data-studio-add="relations-out"]').click();
    const ned = studio(page).locator('[data-studio-editor="meta-relations"]');
    await expect(ned.locator('[data-locked="metaId"]')).toBeVisible();
    await page.locator('[data-studio-back]').first().click();
    await settleStudio(page);

    await page.keyboard.press('Escape');
    await expect(studio(page)).toHaveCount(0);
    await expect(page).not.toHaveURL(/studio=/);

    await page.goto(CONSOLE_URL + '/#/economy/metas?studio=' + M.lvl.id);
    await settleStudio(page);
    await expect(studio(page).locator('[data-studio-group="relations-in"]')).toContainText('بالا رفتن ' + k);
    await studio(page).locator('[data-studio-edit-meta]').click();
    await settlePanel(page);
    await expect(page).toHaveURL(new RegExp('#/economy/metas/' + M.lvl.id));
    await expect(field(panel(page), /^کلید$/).locator('input')).toHaveValue('e2e_' + k + '_lvl');
  });

  test('the studio fits a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await open(page, '#/economy/metas?studio=' + M.xp.id);
    await settleStudio(page);
    const box = await panel(page).locator('[role="dialog"]').boundingBox();
    expect(box!.width).toBeLessThanOrEqual(390);
    const over = await studio(page).evaluate((el) => el.scrollWidth - el.clientWidth);
    expect(over).toBeLessThanOrEqual(1);
  });
});

test.describe('meta relations: player card and policies', () => {
  test('an admin meta change fires a relation, the card lists it, and a reset lets it fire again', async ({ page }) => {
    const rel = await eco.post('/meta-relations', { key: 'e2e_' + k + '_rich', name: { fa: 'ثروتمند ' + k }, metaId: M.gold.id, operator: 'gte', value: 10, targets: [{ metaId: M.lvl.id, action: 'set', amount: 7 }], isActive: true });
    made.rels.push(rel.id);
    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-progress]');
    await expect(card.locator('h2')).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('[data-pp-relations]')).toContainText('هنوز رابطه‌ای');
    await card.locator('[data-change-meta]').click();
    const dlg = page.locator('[data-form-dialog="change-meta"]');
    await expect(dlg).toBeVisible();
    await pickIn(dlg, 'e2e_' + k + '_gold');
    await dlg.locator('input[inputmode="decimal"]').fill('۱۲');
    await dlg.locator('[data-dialog-ok]').click();
    await expect(toasts(page).last()).toContainText('متا تغییر کرد');
    await expect(card.locator(`[data-pp-relation="e2e_${k}_rich"]`)).toBeVisible();
    const lvl = sql(`SELECT value::int FROM economy_user_meta WHERE user_id='${P.id}' AND meta_id='${M.lvl.id}'`);
    expect(lvl).toBe('7');
    await card.locator(`[data-relation-reset="e2e_${k}_rich"]`).click();
    await page.locator('[data-economy="dialog"] button', { hasText: 'از نو' }).click();
    await expect(toasts(page).last()).toContainText('از نو شد');
    await expect(card.locator(`[data-pp-relation="e2e_${k}_rich"]`)).toHaveCount(0);
    expect(sql(`SELECT count(*) FROM economy_user_meta_relation WHERE user_id='${P.id}' AND relation_id='${rel.id}'`)).toBe('0');
  });

  test('the chain depth and empty prize group policies are saved in the event settings', async ({ page }) => {
    await openSettings(page, 'events');
    const body = psWindow(page);
    await expect(body).toContainText('رابطهٔ متاها');
    await field(body, 'عمق زنجیرهٔ رابطه‌ها').locator('input').fill('۳۰');
    await body.locator('[data-settings-save]').click();
    await expect(field(body, 'عمق زنجیرهٔ رابطه‌ها')).toContainText('عدد');
    await field(body, 'عمق زنجیرهٔ رابطه‌ها').locator('input').fill('۳');
    await field(body, 'وقتی گروه جایزه خالی است').locator('select').selectOption('wait');
    await body.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    expect(await eco.get('/event-settings')).toMatchObject({ relationChainDepth: 3, relationPrizeMissing: 'wait', timezone: settings.timezone });
  });
});
