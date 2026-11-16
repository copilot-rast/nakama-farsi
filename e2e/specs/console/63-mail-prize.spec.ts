import { expect, test } from '../../lib/fixtures';
import { open, root } from '../../lib/console';
import { consoleApi, deleteUser, devicePlayer, pub, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { API_URL } from '../../lib/env';

function errCode(res: { data: any }) {
  try { return JSON.parse(res.data.message).code; } catch { return res.data && res.data.message; }
}

async function wallet(p: Player): Promise<Record<string, number>> {
  const r = await fetch(API_URL + '/v2/account', { headers: { Authorization: 'Bearer ' + p.token } });
  return JSON.parse((await r.json()).wallet || '{}');
}

test.describe('in-app mail with an attached prize', () => {
  const players: Player[] = [];
  test.afterAll(async () => {
    for (const p of players) await deleteUser(p.id);
  });

  test('a message carries a prize; each player claims it once from the inbox', async ({ page }) => {
    const prize = sql(`SELECT id::text || '|' || (items->0->>'key') || '|' || (items->0->>'value') FROM economy_prize WHERE is_repeatable AND jsonb_array_length(items)=1 AND items->0->>'type'='currency' AND items->0->>'action'='add' ORDER BY key LIMIT 1`).split('|');
    expect(prize.length, 'a repeatable single-currency prize must exist').toBe(3);
    const [prizeId, currency, amount] = prize;
    const a = await devicePlayer('e2e-mail-a-' + Date.now());
    const b = await devicePlayer('e2e-mail-b-' + Date.now());
    players.push(a, b);
    const n = await pub.post('/notifications', { channel: 'in_app', title: { fa: 'هدیه' }, body: { fa: 'جایزه‌ات را بگیر' }, prizeId, target: { mode: 'users', userIds: [a.id, b.id] } });
    const bad = await consoleApi('POST', '/v2/console/publication/notifications', { channel: 'sms', body: { fa: 'x' }, prizeId, target: { mode: 'users', userIds: [a.id] } });
    expect(JSON.stringify(bad.data)).toContain('PRIZE_IN_APP_ONLY');

    let item: any;
    await expect.poll(async () => {
      item = ((await rpc(a, 'publication_inbox', { lang: 'fa' })).data.items || []).find((x: any) => x.id === n.id);
      return !!item;
    }, { timeout: 30_000 }).toBe(true);
    expect(item.prize.id).toBe(prizeId);
    expect(item.claimed).toBe(false);

    const before = (await wallet(a))[currency] || 0;
    const claim = await rpc(a, 'publication_inbox_claim', { notificationId: n.id });
    expect(claim.status).toBe(200);
    expect(claim.data.claimed).toBe(true);
    expect((await wallet(a))[currency]).toBe(before + Number(amount));
    expect(errCode(await rpc(a, 'publication_inbox_claim', { notificationId: n.id }))).toBe('ALREADY_CLAIMED');
    const again = ((await rpc(a, 'publication_inbox', { lang: 'fa' })).data.items || []).find((x: any) => x.id === n.id);
    expect(again.claimed).toBe(true);

    const bBefore = (await wallet(b))[currency] || 0;
    expect((await rpc(b, 'publication_inbox_claim', { notificationId: n.id })).status).toBe(200);
    expect((await wallet(b))[currency]).toBe(bBefore + Number(amount));
    expect(sql(`SELECT count(*) FROM economy_user_prize WHERE source_type='mail' AND source_id LIKE '${n.id}:%'`)).toBe('2');

    const plain = await pub.post('/notifications', { channel: 'in_app', title: { fa: 'خبر' }, body: { fa: 'بدون جایزه' }, target: { mode: 'users', userIds: [a.id] } });
    await expect.poll(async () => ((await rpc(a, 'publication_inbox', {})).data.items || []).some((x: any) => x.id === plain.id), { timeout: 30_000 }).toBe(true);
    expect(errCode(await rpc(a, 'publication_inbox_claim', { notificationId: plain.id }))).toBe('MESSAGE_HAS_NO_PRIZE');
    expect(errCode(await rpc(b, 'publication_inbox_claim', { notificationId: plain.id }))).toBe('MESSAGE_NOT_FOUND');

    await open(page, '#/publication/notifications/new');
    await expect(root(page).locator('[data-nt-prize-pick]')).toBeVisible();
    await root(page).locator('[data-nt-channels] [data-value="sms"]').click();
    await expect(root(page).locator('[data-nt-prize-pick]')).toBeHidden();
  });
});
