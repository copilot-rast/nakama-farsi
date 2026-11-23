import { expect, test } from '../../lib/fixtures';
import { codeCount, otpSandboxOn, removePhones, resetOtpLimits, testPhone, waitCode } from '../../lib/accounts';
import { consoleApi, deleteUser, devicePlayer, must, rpc, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { API_URL } from '../../lib/env';

const made: string[] = [];

async function nakama(player: Player, path: string, init: RequestInit = {}) {
  const r = await fetch(API_URL + path, { ...init, headers: { Authorization: 'Bearer ' + player.token, 'Content-Type': 'application/json' } });
  return { status: r.status, text: await r.text() };
}

function digitsOf(e164: string) {
  return e164.slice(3);
}

test.describe.configure({ mode: 'serial' });

test.describe('phone numbers stay private', () => {
  test.afterAll(async () => {
    for (const id of made) await deleteUser(id);
    await removePhones();
  });

  test('a phone linked by SMS is not in metadata and other players never see it', async () => {
    expect(otpSandboxOn(), 'SMS login must be in sandbox before the e2e run').toBe(true);
    resetOtpLimits();
    const phone = testPhone();
    const digits = digitsOf(phone.e164);
    const a = await devicePlayer('e2e-pp-a-' + Date.now());
    const b = await devicePlayer('e2e-pp-b-' + Date.now());
    made.push(a.id, b.id);
    const before = codeCount(phone.e164);
    expect((await rpc(a, 'publication_otp_request', { phone: phone.local })).status).toBe(200);
    const code = await waitCode(phone.e164, before);
    const verify = await rpc(a, 'publication_otp_verify', { phone: phone.e164, code, link: true });
    expect(verify.status).toBe(200);
    expect(verify.data.userId).toBe(a.id);
    expect(sql(`SELECT phone FROM publication_phone_account WHERE user_id='${a.id}'`)).toBe(phone.e164);
    expect(sql(`SELECT metadata ? 'phone' FROM users WHERE id='${a.id}'`)).toBe('f');

    const own = await rpc(a, 'publication_phone', {});
    expect(own.status).toBe(200);
    expect(own.data.phone).toBe(phone.e164);
    const others = await rpc(b, 'publication_phone', {});
    expect(others.data.phone).toBe('');

    const checks: [string, string][] = [];
    const users = await rpc(b, 'base_users_get', { ids: [a.id] });
    expect(users.status).toBe(200);
    expect(users.data.users).toHaveLength(1);
    checks.push(['base_users_get', JSON.stringify(users.data)]);
    checks.push(['GET /v2/user', (await nakama(b, '/v2/user?ids=' + a.id)).text]);
    checks.push(['base_friends_add', JSON.stringify((await rpc(b, 'base_friends_add', { ids: [a.id] })).data)]);
    checks.push(['base_friends_add accept', JSON.stringify((await rpc(a, 'base_friends_add', { ids: [b.id] })).data)]);
    const friends = await rpc(b, 'base_friends_list', {});
    expect(friends.status).toBe(200);
    expect(friends.data.friends.map((f: any) => f.user.id)).toContain(a.id);
    checks.push(['base_friends_list', JSON.stringify(friends.data)]);
    checks.push(['GET /v2/friend', (await nakama(b, '/v2/friend')).text]);
    const myCode = await rpc(a, 'base_my_code', {});
    checks.push(['base_user_by_code', JSON.stringify((await rpc(b, 'base_user_by_code', { code: myCode.data.code })).data)]);
    checks.push(['base_player_codes', JSON.stringify((await rpc(b, 'base_player_codes', { ids: [a.id] })).data)]);
    checks.push(['base_friend_suggestions', JSON.stringify((await rpc(b, 'base_friend_suggestions', {})).data)]);
    checks.push(['base_account_get', JSON.stringify((await rpc(a, 'base_account_get', {})).data)]);
    for (const [what, body] of checks) {
      expect(body, what).not.toContain(digits);
      expect(body, what).not.toMatch(/\\?"phone\\?"/);
    }

    await must(consoleApi('POST', '/v2/console/account/' + a.id, { metadata: JSON.stringify({ phone: phone.e164, mobile: phone.local, country: 'IR' }) }));
    const stale = await rpc(b, 'base_users_get', { ids: [a.id] });
    expect(stale.status).toBe(200);
    expect(JSON.stringify(stale.data)).not.toContain(digits);
    expect(JSON.parse(stale.data.users[0].metadata)).toEqual({ country: 'IR' });
    const list = await rpc(b, 'base_friends_list', {});
    expect(JSON.stringify(list.data)).not.toContain(digits);

    const card = await must(consoleApi('GET', '/v2/console/base/players/' + a.id));
    expect(card.phone).toBe(phone.e164);
    const found = await must(consoleApi('GET', '/v2/console/base/players?q=' + encodeURIComponent(phone.local)));
    expect(found.items.map((x: any) => x.id)).toEqual([a.id]);
  });
});
