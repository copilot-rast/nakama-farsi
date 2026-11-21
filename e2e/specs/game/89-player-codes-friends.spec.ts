import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { Chat, notice } from '../../lib/chat';
import { API_URL } from '../../lib/env';
import { sql } from '../../lib/db';

const k = uid();
const users: string[] = [];
const sockets: Chat[] = [];
let A: Player;
let B: Player;
let C: Player;
let savedRules: any;

test.describe.configure({ mode: 'serial' });

async function socket(p: Player) {
  const c = await Chat.open(p, { status: true });
  sockets.push(c);
  return c;
}

async function inbox(p: Player, code: number) {
  const r = await fetch(API_URL + '/v2/notification?limit=100', { headers: { Authorization: 'Bearer ' + p.token } });
  const body: any = await r.json();
  return (body.notifications || []).filter((n: any) => Number(n.code) === code).map((n: any) => ({ subject: n.subject, content: JSON.parse(n.content || '{}'), sender: n.sender_id }));
}

test.beforeAll(async () => {
  savedRules = await base.get('/friend-rules');
  A = await devicePlayer('e2e-pc-a-' + k, ('pca' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-pc-b-' + k, ('pcb' + k).replace(/_/g, '').slice(0, 20));
  C = await devicePlayer('e2e-pc-c-' + k, ('pcc' + k).replace(/_/g, '').slice(0, 20));
  users.push(A.id, B.id, C.id);
  await updateAccount(A, { lang_tag: 'fa', display_name: 'آرش ' + k.slice(-4) });
  await updateAccount(B, { lang_tag: 'en', display_name: 'Bita ' + k.slice(-4) });
  await base.put('/friend-rules', { ...savedRules, notifyDeclined: true, notifyRemoved: true, presenceFriendsOnly: false, dmFriendsOnly: false });
});

test.afterAll(async () => {
  for (const s of sockets) s.close();
  const { updatedBy, updatedAt, ...rules } = savedRules;
  await base.put('/friend-rules', rules);
  for (const id of users) await deleteUser(id);
});

test.describe('game: player codes, friend notices and friends-only rules', () => {
  test('every account has a code; lookup by code over the socket and HTTP; friends added by code', async () => {
    const a = await socket(A);
    const mine = await a.rpc('base_my_code', {});
    const codeA = mine.data.code as string;
    expect(codeA).toMatch(/^[A-Z0-9]{6}$/);
    expect(sql(`SELECT code FROM base_player_code WHERE user_id='${A.id}'`)).toBe(codeA);
    const codes = await rpc(A, 'base_player_codes', { ids: [B.id, C.id] });
    expect(Object.keys(codes.data.codes).sort()).toEqual([B.id, C.id].sort());

    const b = await socket(B);
    const found = await b.rpc('base_user_by_code', { code: codeA.toLowerCase() });
    expect(found.data).toMatchObject({ user: { id: A.id, code: codeA, displayName: 'آرش ' + k.slice(-4) }, relationship: 'none' });
    expect(found.data.online).toBeUndefined();
    const missing = await b.rpc('base_user_by_code', { code: 'QQQQQQQQQQ' });
    expect(missing.error?.message).toContain('PLAYER_CODE_NOT_FOUND');

    const sent = await b.rpc('base_friends_add', { codes: [codeA] });
    expect(sent.error).toBeUndefined();
    expect(sent.data.friends[0]).toMatchObject({ state: 1, user: { id: A.id } });
    const own = await b.rpc('base_friends_add', { codes: [(await b.rpc('base_my_code', {})).data.code] });
    expect(own.error?.message).toContain('CANNOT_TARGET_SELF');
    expect((await b.rpc('base_user_by_code', { code: codeA })).data.relationship).toBe('sent');
  });

  test('declining a request and removing a friend leave an inbox notice for the other player', async () => {
    const a = await socket(A);
    const declined = await a.rpc('base_friends_remove', { ids: [B.id] });
    expect(declined.error).toBeUndefined();
    await expect.poll(async () => (await inbox(B, 1130)).length).toBe(1);
    const [n] = await inbox(B, 1130);
    expect(n.content).toMatchObject({ code: 'FRIEND_REQUEST_DECLINED', userId: A.id });
    expect(n.subject).toContain('declined your friend request');
    expect(n.sender).toBe(A.id);

    const b = await socket(B);
    await b.rpc('base_friends_add', { ids: [A.id] });
    expect((await a.rpc('base_friends_add', { ids: [B.id] })).error).toBeUndefined();
    expect((await a.rpc('base_user_by_code', { code: sql(`SELECT code FROM base_player_code WHERE user_id='${B.id}'`) })).data).toMatchObject({ relationship: 'friends', online: true });
    expect((await b.rpc('base_friends_remove', { ids: [A.id] })).error).toBeUndefined();
    await expect.poll(async () => (await inbox(A, 1131)).length).toBe(1);
    const [r] = await inbox(A, 1131);
    expect(r.content).toMatchObject({ code: 'FRIEND_REMOVED', userId: B.id });
    expect(r.subject).toContain('از دوستانش حذف کرد');
    expect(await inbox(B, 1131)).toHaveLength(0);
  });

  test('with friends-only rules a stranger cannot open a private chat and sees no presence', async () => {
    const a = await socket(A);
    const b = await socket(B);
    const c = await socket(C);
    await b.rpc('base_friends_add', { ids: [A.id] });
    await a.rpc('base_friends_add', { ids: [B.id] });
    expect(new Set((await c.follow([A.id, B.id])).presences.map((p: any) => p.user_id)).size).toBe(2);
    await base.put('/friend-rules', { ...savedRules, updatedBy: undefined, updatedAt: undefined, presenceFriendsOnly: true, dmFriendsOnly: true });
    expect((await c.rpc('base_friend_rules', {})).data).toMatchObject({ presenceFriendsOnly: true, dmFriendsOnly: true });

    let err = '';
    try { await c.joinDirect(A.id); } catch (e: any) { err = String(e.message); }
    expect(notice(err.replace(/^join failed: /, ''))).toMatchObject({ code: 'FRIENDSHIP_REQUIRED', meta: { userId: A.id } });
    await a.joinDirect(B.id);
    expect((await a.send('سلام ' + k)).ok).toBe(true);

    const seen = await c.follow([A.id, B.id]);
    expect(seen.presences).toHaveLength(0);
    const friendView = await a.follow([B.id, C.id]);
    expect([...new Set(friendView.presences.map((p: any) => p.user_id))]).toEqual([B.id]);

    await b.rpc('base_friends_remove', { ids: [A.id] });
    const stale = await a.send('هنوز؟');
    expect(stale.ok).toBe(false);
    expect(notice((stale as any).message).code).toBe('FRIENDSHIP_REQUIRED');
  });
});
