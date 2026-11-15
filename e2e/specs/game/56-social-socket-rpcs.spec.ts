import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { Chat } from '../../lib/chat';
import { sql } from '../../lib/db';

// Friends, own account and user lookup as RPCs (integrations/base/social_rpc.go),
// sent over the socket the way the Rakhtkan app's callRpc sends them, with the
// HTTP fallback giving the same answer. The friend rules of the HTTP friend API
// (daily limit, cooldown after a decline) apply to the RPCs too.

const k = uid();
const users: string[] = [];
const sockets: Chat[] = [];
let before: any;
let A: Player;
let B: Player;
let C: Player;
let sa: Chat;
let sb: Chat;
let sc: Chat;

test.describe.configure({ mode: 'serial' });

async function player(name: string): Promise<Player> {
  const p = await devicePlayer('e2e-so-' + name + '-' + k, ('so' + name + k).replace(/_/g, '').slice(0, 24));
  users.push(p.id);
  await updateAccount(p, { lang_tag: 'fa' });
  return p;
}

async function socket(p: Player) {
  const c = await Chat.open(p);
  sockets.push(c);
  return c;
}

function states(out: any): Record<string, number> {
  const res: Record<string, number> = {};
  for (const f of out?.friends || []) res[f.user.id] = f.state ?? 0;
  return res;
}

function code(err: any): string {
  try { return JSON.parse(err.message).code; } catch { return String(err?.message || ''); }
}

test.beforeAll(async () => {
  before = await base.get('/friend-rules');
  await base.put('/friend-rules', { dailyRequests: 1, maxFriends: 0, expiryDays: 0, rejectCooldownHours: 24 });
  A = await player('a');
  B = await player('b');
  C = await player('c');
  sa = await socket(A);
  sb = await socket(B);
  sc = await socket(C);
});

test.afterAll(async () => {
  for (const s of sockets) s.close();
  await base.put('/friend-rules', { dailyRequests: before.dailyRequests, maxFriends: before.maxFriends, expiryDays: before.expiryDays, rejectCooldownHours: before.rejectCooldownHours, requestsBonusMeta: before.requestsBonusMeta || '', friendsBonusMeta: before.friendsBonusMeta || '', matchPolicy: before.matchPolicy || 'priority', overLimit: before.overLimit || 'refuse' });
  sql(`UPDATE base_friend_rules SET updated_by='${String(before.updatedBy || '').replace(/'/g, "''")}', updated_at='${before.updatedAt}' WHERE id=1`);
  sql(`DELETE FROM base_friend_request_day WHERE user_id IN (${users.map((u) => "'" + u + "'").join(',')})`);
  sql(`DELETE FROM base_friend_rejection WHERE requester_id IN (${users.map((u) => "'" + u + "'").join(',')}) OR rejecter_id IN (${users.map((u) => "'" + u + "'").join(',')})`);
  for (const id of users) await deleteUser(id);
});

test('account and profile over the socket, same answer on HTTP', async () => {
  const name = ('sop' + k).replace(/_/g, '').slice(0, 20);
  const upd = await sa.rpc('base_account_update', { username: name, display_name: 'سامان' });
  expect(upd.error).toBeUndefined();
  expect(upd.data.user.username).toBe(name);
  expect(upd.data.user.display_name).toBe('سامان');
  const viaSocket = await sa.rpc('base_account_get', {});
  const viaHttp = await rpc(A, 'base_account_get', {});
  expect(viaHttp.status).toBe(200);
  expect(viaSocket.data.user.id).toBe(A.id);
  expect(viaHttp.data.user).toEqual(viaSocket.data.user);
  expect(viaSocket.data.disable_time).toBeUndefined();

  const taken = await sb.rpc('base_account_update', { username: name });
  expect(taken.error?.message).toBe('USERNAME_TAKEN');

  const found = await sb.rpc('base_users_get', { usernames: [name] });
  expect(found.data.users.map((u: any) => u.id)).toEqual([A.id]);
  const none = await sb.rpc('base_users_get', { usernames: ['nobody' + k] });
  expect(none.data.users ?? []).toEqual([]);
});

test('request, accept, daily limit, decline cooldown, block and unblock over the socket', async () => {
  const sent = await sa.rpc('base_friends_add', { ids: [B.id] });
  expect(sent.error).toBeUndefined();
  expect(states(sent.data)[B.id]).toBe(1);

  const limited = await sa.rpc('base_friends_add', { ids: [C.id], lang: 'fa' });
  expect(code(limited.error)).toBe('FRIEND_DAILY_LIMIT');
  expect(JSON.parse(limited.error!.message).lang).toBe('fa');
  expect(JSON.parse(limited.error!.message).message).toContain('۱');

  const incoming = await sb.rpc('base_friends_list', { state: 2 });
  expect(states(incoming.data)[A.id]).toBe(2);
  const accepted = await sb.rpc('base_friends_add', { ids: [A.id] });
  expect(states(accepted.data)[A.id]).toBe(0);
  const list = await sa.rpc('base_friends_list', { state: 0, limit: 10 });
  expect(states(list.data)[B.id]).toBe(0);
  expect(list.data.friends[0].user.username).toBeTruthy();
  expect(list.data.friends[0].update_time).toBeTruthy();

  const fromC = await sc.rpc('base_friends_add', { ids: [B.id] });
  expect(states(fromC.data)[B.id]).toBe(1);
  const declined = await sb.rpc('base_friends_remove', { ids: [C.id] });
  expect(states(declined.data)[C.id]).toBeUndefined();
  sql(`DELETE FROM base_friend_request_day WHERE user_id='${C.id}'`);
  const again = await sc.rpc('base_friends_add', { ids: [B.id] });
  expect(code(again.error)).toBe('FRIEND_REQUEST_COOLDOWN');

  const blocked = await sb.rpc('base_friends_block', { ids: [C.id] });
  expect(states(blocked.data)[C.id]).toBe(3);
  const blockList = await rpc(B, 'base_friends_list', { state: 3 });
  expect(blockList.status).toBe(200);
  expect(states(blockList.data)[C.id]).toBe(3);
  const unblocked = await sb.rpc('base_friends_remove', { ids: [C.id] });
  expect(states(unblocked.data)[C.id]).toBeUndefined();

  const self = await sa.rpc('base_friends_add', { ids: [A.id] });
  expect(self.error?.message).toBe('CANNOT_TARGET_SELF');
});
