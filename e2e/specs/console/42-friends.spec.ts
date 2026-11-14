import { expect, test } from '../../lib/fixtures';
import { API_URL } from '../../lib/env';
import { base, deleteUser, devicePlayer, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, openSettings, root, settle, toasts } from '../../lib/console';

const k = uid();
const users: string[] = [];
let before: any;

test.describe.configure({ mode: 'serial' });

async function player(name: string, lang = 'fa'): Promise<Player> {
  const p = await devicePlayer('e2e-fr-' + name + '-' + k, ('fr' + name + k).replace(/_/g, '').slice(0, 30));
  users.push(p.id);
  await updateAccount(p, { lang_tag: lang });
  return p;
}

async function friendCall(method: string, p: Player, ids: string[]) {
  const r = await fetch(API_URL + '/v2/friend?' + ids.map((id) => 'ids=' + encodeURIComponent(id)).join('&'), { method, headers: { Authorization: 'Bearer ' + p.token } });
  const text = await r.text();
  let data: any = null;
  try { data = JSON.parse(text); } catch { data = text; }
  let notice: any = null;
  try { notice = data && data.message ? JSON.parse(data.message) : null; } catch { notice = null; }
  return { status: r.status, data, notice };
}

const add = (p: Player, ...ids: string[]) => friendCall('POST', p, ids);
const remove = (p: Player, ...ids: string[]) => friendCall('DELETE', p, ids);

test.beforeAll(async () => {
  before = await base.get('/friend-rules');
});

test.afterAll(async () => {
  await base.put('/friend-rules', { dailyRequests: before.dailyRequests, maxFriends: before.maxFriends, expiryDays: before.expiryDays, rejectCooldownHours: before.rejectCooldownHours });
  sql(`UPDATE base_friend_rules SET updated_by='${String(before.updatedBy || '').replace(/'/g, "''")}', updated_at='${before.updatedAt}' WHERE id=1`);
  for (const id of users) await deleteUser(id);
});

test.describe('friends: rules in project settings', () => {
  test('defaults, summary sentence, limits checked in the form, save and the audit log', async ({ page }) => {
    await base.put('/friend-rules', { dailyRequests: 20, maxFriends: 0, expiryDays: 0, rejectCooldownHours: 0 });
    await openSettings(page, 'friends');
    const form = root(page).locator('[data-settings-form="friends"]');
    await expect(form.locator('form')).toBeVisible();
    await expect(form).toContainText('درخواست در روز');
    const daily = field(form, 'درخواست در روز').locator('input');
    const most = field(form, 'بیشترین تعداد دوست').locator('input');
    const expiry = field(form, 'انقضای درخواست (روز)').locator('input');
    const cooldown = field(form, 'فاصله پس از رد شدن (ساعت)').locator('input');
    await expect(daily).toHaveValue(/^(20|۲۰)$/);
    const summary = form.locator('[data-friend-summary]');
    await expect(summary).toContainText('روزی ۲۰ درخواست');
    await expect(summary).toContainText('تعداد دوست بی‌حد');
    await expect(summary).toContainText('درخواست‌ها منقضی نمی‌شوند');
    await expect(summary).toContainText('بدون فاصله پس از رد شدن');

    await daily.fill('2000');
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'درخواست در روز')).toContainText('یک عدد معتبر وارد کنید');
    expect((await base.get('/friend-rules')).dailyRequests).toBe(20);
    await daily.fill('3');
    await most.fill('5');
    await expiry.fill('7');
    await cooldown.fill('24');
    await expect(summary).toContainText('روزی ۳ درخواست');
    await expect(summary).toContainText('تا ۵ دوست');
    await expect(summary).toContainText('پس از ۷ روز پاک می‌شود');
    await expect(summary).toContainText('۲۴ ساعت صبر');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    await settle(page);
    const saved = await base.get('/friend-rules');
    expect(saved).toMatchObject({ dailyRequests: 3, maxFriends: 5, expiryDays: 7, rejectCooldownHours: 24 });
    await expect(form).toContainText('آخرین ذخیره');
    const audit = sql(`SELECT count(*) FROM console_audit_log WHERE message LIKE 'base update friend-rules%' AND create_time > now() - interval '5 minutes'`);
    expect(Number(audit)).toBeGreaterThan(0);
  });
});

test.describe('friends: rules on the server', () => {
  test('daily limit with a translated JSON error, remaining count, accepting is free, max friends, cooldown after a rejection', async () => {
    await base.put('/friend-rules', { dailyRequests: 2, maxFriends: 2, expiryDays: 0, rejectCooldownHours: 24 });
    const a = await player('a');
    const [b, c, d, e] = [await player('b'), await player('c', 'en'), await player('d'), await player('e')];

    let rules = (await rpc(a, 'base_friend_rules', {})).data;
    expect(rules).toMatchObject({ dailyRequests: 2, sentToday: 0, remainingToday: 2, maxFriends: 2, friends: 0, rejectCooldownHours: 24 });
    expect(Date.parse(rules.resetsAt)).toBeGreaterThan(Date.now());

    expect((await add(a, b.id)).status).toBe(200);
    expect((await add(a, c.id)).status).toBe(200);
    rules = (await rpc(a, 'base_friend_rules', {})).data;
    expect(rules).toMatchObject({ sentToday: 2, remainingToday: 0 });
    const over = await add(a, d.id);
    expect(over.status).toBe(429);
    expect(over.notice).toMatchObject({ code: 'FRIEND_DAILY_LIMIT', lang: 'fa', limit: 2, remaining: 0 });
    expect(over.notice.message).toContain('امروز بیشتر از ۲ درخواست دوستی');
    expect(over.notice.resetsAt).toBe(rules.resetsAt);

    expect((await add(b, a.id)).status).toBe(200);
    expect((await rpc(b, 'base_friend_rules', {})).data).toMatchObject({ sentToday: 0, friends: 1 });

    const cDecline = await remove(c, a.id);
    expect(cDecline.status).toBe(200);
    sql(`UPDATE base_friend_request_day SET count = 0 WHERE user_id='${a.id}'`);
    const again = await add(a, c.id);
    expect(again.status).toBe(400);
    expect(again.notice).toMatchObject({ code: 'FRIEND_REQUEST_COOLDOWN', userId: c.id });
    expect(Date.parse(again.notice.retryAt) - Date.now()).toBeGreaterThan(23 * 3600_000);

    const ce = await add(c, a.id);
    expect(ce.status).toBe(200);
    expect((await add(a, c.id)).status).toBe(200);
    expect((await rpc(a, 'base_friend_rules', {})).data).toMatchObject({ friends: 2 });

    const full = await add(a, e.id);
    expect(full.status).toBe(400);
    expect(full.notice).toMatchObject({ code: 'FRIEND_LIMIT_REACHED', limit: 2 });
    expect(full.notice.message).toContain('فهرست دوستانت پر است');
    expect((await add(e, a.id)).status).toBe(200);
    const theirs = await add(d, e.id);
    expect(theirs.status).toBe(200);
    expect((await add(e, d.id)).status).toBe(200);
    expect((await add(e, a.id)).status).toBe(200);
    const accept = await add(a, e.id);
    expect(accept.notice).toMatchObject({ code: 'FRIEND_LIMIT_REACHED' });
  });

  test('expired invites disappear when the list is read and the English message', async () => {
    await base.put('/friend-rules', { dailyRequests: 1, maxFriends: 0, expiryDays: 3, rejectCooldownHours: 0 });
    const x = await player('x', 'en');
    const y = await player('y');
    expect((await add(x, y.id)).status).toBe(200);
    const en = await add(x, (await player('z')).id);
    expect(en.notice).toMatchObject({ code: 'FRIEND_DAILY_LIMIT', lang: 'en' });
    expect(en.notice.message).toContain('at most 1 friend requests a day');
    sql(`UPDATE user_edge SET update_time = now() - interval '4 days' WHERE source_id='${x.id}' AND destination_id='${y.id}'`);
    const r = await fetch(API_URL + '/v2/friend?limit=10', { headers: { Authorization: 'Bearer ' + y.token } });
    const list = await r.json();
    expect(list.friends || []).toHaveLength(0);
    expect(sql(`SELECT count(*) FROM user_edge WHERE source_id IN ('${x.id}','${y.id}')`)).toBe('0');
  });
});
