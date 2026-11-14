import { expect, test } from '../../lib/fixtures';
import type { Locator, Page } from '@playwright/test';
import { CONSOLE_URL } from '../../lib/env';
import { deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, go, noHorizontalScroll, open, panel, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const made = { users: [] as string[], prizes: [] as string[], types: [] as string[], tasks: [] as string[], groups: [] as string[] };
let prize: any;
let eventKey: string;
let P: Player;

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

async function confirm(page: Page, label: string) {
  await page.locator('[role="alertdialog"] button', { hasText: label }).click();
}

test.beforeAll(async () => {
  prize = await eco.post('/prizes', { key: 'e2e_' + k + '_coins', name: { fa: 'جایزهٔ تسک ' + k, en: 'Task prize ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 9 }], tagIds: [], metadata: {} });
  made.prizes.push(prize.id);
  eventKey = 'e2e_' + k + '_hit';
  const et = await eco.post('/event-types', { key: eventKey, name: { fa: 'ضربهٔ آزمایشی ' + k, en: 'Test hit ' + k }, description: {}, clientAllowed: true, valueMin: 0, valueMax: 50, rateLimit: 100, isActive: true, sort: 0 });
  made.types.push(et.id);
  P = await devicePlayer('e2e-tk-' + k, ('tk' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa' });
});

test.afterAll(async () => {
  for (const id of made.users) await deleteUser(id);
  sql(`DELETE FROM economy_reward_claim WHERE prize_id='${prize.id}'; DELETE FROM economy_user_prize WHERE prize_id='${prize.id}'`);
  const all = await eco.get('/tasks?q=e2e_' + k + '&limit=100');
  const ids = (all.items || []).sort((a: any, b: any) => (a.parentId ? 0 : 1) - (b.parentId ? 0 : 1)).map((x: any) => x.id);
  for (const id of ids) await eco.del('/tasks/' + id);
  for (const id of made.groups) await eco.del('/task-groups/' + id);
  for (const id of made.types) await eco.del('/event-types/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
});

test.describe('tasks: console', () => {
  test('task editor: live condition sentence, validation, save, list sentence, search and the sub-task filter', async ({ page }) => {
    await open(page, '#/economy/tasks');
    await expect(root(page)).toContainText('کارهایی که بازیکن در بازی انجام می‌دهد');
    await go(page, '#/economy/tasks/new');
    await settlePanel(page);
    let p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_daily');
    await fillTranslated(p, 'name', { fa: 'سه ضربه در روز ' + k, en: 'Three hits a day ' + k });
    await pickIn(field(p, 'رویداد'), 'ضربهٔ آزمایشی ' + k);
    await field(p, 'هدف').locator('input').fill('۳');
    await field(p, 'کمترین مقدار قابل‌قبول').locator('input').fill('5');
    await field(p, 'تکرار').locator('select').selectOption('daily');
    await expect(p.locator('[data-task-preview]')).toHaveText(`«ضربهٔ آزمایشی ${k}» ۳ بار · فقط مقدار ۵ به بالا · هر روز`);
    await field(p, 'تکرار').locator('select').selectOption('interval');
    await savePanel(page).catch(() => null);
    await expect(field(p, 'فاصله (دقیقه)')).toContainText('لازم');
    await field(p, 'تکرار').locator('select').selectOption('daily');
    await field(p, 'بیشترین مقدار قابل‌قبول').locator('input').fill('2');
    await savePanel(page).catch(() => null);
    await expect(field(p, 'بیشترین مقدار قابل‌قبول')).toContainText('کوچک‌تر');
    await field(p, 'بیشترین مقدار قابل‌قبول').locator('input').fill('');
    await pickIn(field(p, 'جایزه'), 'جایزهٔ تسک ' + k);
    await savePanel(page);
    const daily = await byKey('tasks', 'e2e_' + k + '_daily');
    made.tasks.push(daily.id);
    expect(daily).toMatchObject({ condition: 'event', metric: 'count', target: 3, valueMin: 5, repeat: 'daily', grantMode: 'claim', prizeId: prize.id });
    await search(page, 'e2e_' + k + '_daily');
    await expect(rows(page)).toHaveCount(1);
    await expect(rows(page).first().locator('[data-task-sentence]')).toHaveText(`«ضربهٔ آزمایشی ${k}» ۳ بار · فقط مقدار ۵ به بالا · هر روز`);
    await expect(rows(page).first()).toContainText('جایزهٔ تسک ' + k);

    await go(page, '#/economy/tasks/new');
    await settlePanel(page);
    p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_combo');
    await fillTranslated(p, 'name', { fa: 'ترکیبی ' + k });
    await field(p, 'نوع شرط').locator('select').selectOption('children');
    await expect(field(p, 'رویداد')).toBeHidden();
    await expect(p.locator('[data-task-preview]')).toContainText('وقتی همهٔ زیرتسک‌ها انجام شوند');
    await pickIn(field(p, 'جایزه'), 'جایزهٔ تسک ' + k);
    await field(p, 'روش دادن').locator('select').selectOption('auto');
    await savePanel(page);
    const combo = await byKey('tasks', 'e2e_' + k + '_combo');
    made.tasks.push(combo.id);
    const sub = await eco.post('/tasks', { key: 'e2e_' + k + '_combo_a', name: { fa: 'زیرتسک الف ' + k }, eventTypeId: daily.eventTypeId, target: 1, parentId: combo.id, isActive: true });
    await eco.post('/tasks', { key: 'e2e_' + k + '_combo_b', name: { fa: 'زیرتسک ب ' + k }, eventTypeId: daily.eventTypeId, target: 5, parentId: combo.id, isActive: true });
    expect(sub.repeat).toBe('once');

    await go(page, '#/economy/tasks?level=sub&q=e2e_' + k);
    await expect(rows(page)).toHaveCount(2);
    await expect(rows(page).first()).toContainText('ترکیبی ' + k);
    await go(page, '#/economy/tasks?level=top&q=e2e_' + k);
    await expect(rows(page).filter({ hasText: 'ترکیبی ' + k })).toContainText('۲ زیرتسک');
    await rows(page).filter({ hasText: 'ترکیبی ' + k }).locator('td').first().click();
    await settlePanel(page);
    await panel(page).locator('.x-form > div:last-child').getByRole('button', { name: 'حذف' }).click();
    await panel(page).getByRole('button', { name: 'بله، حذف شود' }).click();
    await expect(panel(page).getByRole('alert')).toContainText('زیرتسک دارد');
    await page.keyboard.press('Escape');
  });

  test('group editor: tasks in the chosen order, hand-out sentence, and a repeating task is refused', async ({ page }) => {
    const one = await eco.post('/tasks', { key: 'e2e_' + k + '_g1', name: { fa: 'گام یک ' + k }, eventTypeId: made.types[0], target: 1, prizeId: prize.id, isActive: true });
    const two = await eco.post('/tasks', { key: 'e2e_' + k + '_g2', name: { fa: 'گام دو ' + k }, eventTypeId: made.types[0], target: 1, prizeId: prize.id, isActive: true });
    await open(page, '#/economy/task-groups/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_grp');
    await fillTranslated(p, 'name', { fa: 'زنجیرهٔ آزمایشی ' + k, en: 'Test chain ' + k });
    await field(p, 'تسک بعدی کی بیاید').locator('select').selectOption('request');
    const picker = field(p, /^تسک‌ها$/);
    await pickIn(picker, 'گام دو ' + k);
    await pickIn(picker, 'گام یک ' + k);
    await pickIn(picker, 'سه ضربه در روز ' + k);
    await savePanel(page).catch(() => null);
    await expect(p).toContainText('تکرار را «یک بار» بگذارید');
    await picker.locator('.eco-chip-x').last().click();
    await savePanel(page);
    const group = await byKey('task-groups', 'e2e_' + k + '_grp');
    made.groups.push(group.id);
    expect(group.taskIds).toEqual([two.id, one.id]);
    await search(page, 'e2e_' + k + '_grp');
    await expect(rows(page).first().locator('[data-group-sentence]')).toHaveText('هم‌زمان ۱ تسک · به ترتیب · تسک بعدی با درخواست بازیکن');
    await expect(rows(page).first()).toContainText('۲ تسک');
  });

  test('player card: progress from events, mark done, give the prize, start over and hand out the next group task', async ({ page }) => {
    for (let i = 0; i < 2; i++) await rpc(P, 'economy_event_record', { key: eventKey, value: 9 });
    await rpc(P, 'economy_event_record', { key: eventKey, value: 1 });
    await page.goto(CONSOLE_URL + '/#/players/' + P.id);
    const card = page.locator('[data-player-tasks]');
    await expect(card.locator('h2')).toHaveText('تسک‌های بازیکن', { timeout: 20_000 });
    const daily = card.locator(`[data-pp-task="e2e_${k}_daily"]`);
    await expect(daily).toHaveAttribute('data-state', 'active');
    await expect(daily.locator('[data-task-progress]')).toHaveText('۲ از ۳');
    await expect(card.locator(`[data-pp-task="e2e_${k}_combo"] [data-task-progress]`)).toHaveText('۱ از ۲ زیرتسک');

    await daily.getByRole('button', { name: 'تمامش کن' }).click();
    await confirm(page, 'تمامش کن');
    await expect(toasts(page).last()).toContainText('تسک تمام شد');
    await expect(daily).toHaveAttribute('data-state', 'completed');
    const before = (await rpc(P, 'economy_balances', {})).data.wallet.coin || 0;
    await daily.getByRole('button', { name: 'جایزه را بده' }).click();
    await expect(toasts(page).last()).toContainText('جایزه به بازیکن رسید');
    await expect(daily).toHaveAttribute('data-state', 'claimed');
    expect((await rpc(P, 'economy_balances', {})).data.wallet.coin).toBe(before + 9);
    const ledger = sql(`SELECT count(*) FROM wallet_ledger WHERE user_id='${P.id}' AND metadata->>'sourceType'='task'`);
    expect(Number(ledger)).toBe(1);

    const sub = card.locator(`[data-pp-task="e2e_${k}_combo_b"]`);
    await expect(sub.locator('[data-task-progress]')).toHaveText('۳ از ۵');
    await sub.getByRole('button', { name: 'از نو' }).click();
    await confirm(page, 'از نو');
    await expect(toasts(page).last()).toContainText('پیشرفت صفر شد');
    await expect(sub.locator('[data-task-progress]')).toHaveText('۰ از ۵');

    const grp = card.locator(`[data-pp-group="e2e_${k}_grp"]`);
    await expect(grp).toContainText('زنجیرهٔ آزمایشی ' + k);
    await grp.locator(`[data-group-next="e2e_${k}_grp"]`).click();
    await expect(toasts(page).last()).toContainText('تسک بعدی داده شد');
    await expect(grp.locator(`[data-pp-task="e2e_${k}_g2"]`)).toBeVisible();
    const mine = await rpc(P, 'economy_task_group_current', { key: 'e2e_' + k + '_grp', lang: 'fa' });
    expect(mine.data.group.tasks.map((x: any) => x.key)).toEqual(['e2e_' + k + '_g2']);
    const claims = await eco.get('/event-claims?userId=' + P.id + '&kind=task');
    expect(claims.items[0]).toMatchObject({ sourceKey: 'e2e_' + k + '_daily', status: 'granted' });
  });

  test('English, 390px cards without sideways scroll, and the page search', async ({ page }) => {
    await open(page, '#/economy/overview');
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('گروه‌های تسک');
    await expect(page.locator('.pl-pal-row').first()).toContainText('گروه‌های تسک');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/#\/economy\/tasks\/groups/);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await go(page, '#/economy/tasks?q=e2e_' + k + '_daily');
    await expect(rows(page).first().locator('[data-task-sentence]')).toHaveText(`"Test hit ${k}" 3 times · only values from 5 · every day`);
    await page.setViewportSize({ width: 390, height: 844 });
    await go(page, '#/economy/tasks?q=e2e_' + k);
    await noHorizontalScroll(page);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click().catch(() => null);
  });
});
