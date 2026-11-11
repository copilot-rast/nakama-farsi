import { expect, test } from '@playwright/test';
import { consoleApi, nodePid, nodes, port } from '../env';
import { joinsOf, leavesOf, player, runTag, sleep } from '../lib';

test.describe.configure({ mode: 'serial' });

const isMatchEnded = (e: any) => (e.notifications || [e]).some((n: any) => n.code === -100 && n.subject === 'match_ended');

test('a killed node: leaves, offline and match_ended reach the other nodes, tickets die with it, the node rejoins', async () => {
  const a = await player(2, 'ka');
  const b = await player(3, 'kb');
  const c = await player(4, 'kc');
  const room = `kroom${runTag}`;
  await a.socket.joinChat(room, 1, false, false);
  await b.socket.joinChat(room, 1, false, false);
  const relayed = await a.socket.createMatch();
  await b.socket.joinMatch(relayed.match_id);
  const hosted = await b.rpc('cle2e_create_match', { tag: `k${runTag}` });
  expect(hosted.match_id).toMatch(/\.node3$/);
  await a.socket.joinMatch(hosted.match_id);
  await a.socket.followUsers([b.id]);
  const pool = `k${runTag}`;
  await b.socket.addMatchmaker(`+properties.pool:${pool}`, 2, 2, { pool });
  await sleep(300);
  const killedAt = Date.now();
  process.kill(nodePid(3), 'SIGKILL');
  const roomLeave = await a.wait('onchannelpresence', (e) => leavesOf(e, b.id), 15_000, killedAt);
  const matchLeave = await a.wait('onmatchpresence', (e) => e.match_id === relayed.match_id && leavesOf(e, b.id), 15_000, killedAt);
  const offline = await a.wait('onstatuspresence', (e) => leavesOf(e, b.id), 15_000, killedAt);
  const ended = await a.wait('onnotification', isMatchEnded, 15_000, killedAt);
  expect(roomLeave && matchLeave && offline && ended).toBeTruthy();
  const times = [roomLeave!, matchLeave!, offline!, ended!].map((e) => e.t - killedAt);
  test.info().annotations.push({ type: 'death-detection-ms', description: times.join(',') });
  expect(Math.max(...times)).toBeLessThan(10_000);
  expect(!!(await a.client.getUsers(a.session, [b.id])).users![0].online).toBe(false);
  await c.socket.addMatchmaker(`+properties.pool:${pool}`, 2, 2, { pool });
  await sleep(4000);
  expect(c.count('onmatchmakermatched')).toBe(0);
  const since = Date.now();
  await b.connect(4);
  expect(await a.wait('onstatuspresence', (e) => joinsOf(e, b.id), 8000, since)).not.toBeNull();
  nodes('start', 3);
  const d = await player(3, 'kd');
  await d.socket.joinChat(room, 1, false, false);
  expect(await a.wait('onchannelpresence', (e) => joinsOf(e, d.id), 8000, since)).not.toBeNull();
  for (const p of [a, b, c, d]) p.close();
});

test('logout on one node invalidates the token on every node; a ban disconnects the socket on another node', async () => {
  const u = await player(1, 'lu');
  expect((await u.rpc('cle2e_whoami')).node).toBe('node1');
  const onNode2 = async () => fetch(`http://127.0.0.1:${port(2)}/v2/rpc/cle2e_whoami`, { method: 'POST', headers: { Authorization: 'Bearer ' + u.session.token }, body: '""' }).then((r) => r.status);
  expect(await onNode2()).toBe(200);
  await sleep(1100);
  await u.rpc('cle2e_session', { op: 'logout', user_id: u.id });
  await expect.poll(onNode2, { timeout: 3000 }).toBe(401);

  const v = await player(3, 'lv');
  const w = await player(1, 'lw');
  await sleep(1100);
  const since = Date.now();
  await w.rpc('cle2e_session', { op: 'ban', user_id: v.id });
  expect(await v.wait('ondisconnect', () => true, 8000, since)).not.toBeNull();
  const onNode4 = await fetch(`http://127.0.0.1:${port(4)}/v2/rpc/cle2e_whoami`, { method: 'POST', headers: { Authorization: 'Bearer ' + v.session.token }, body: '""' }).then((r) => r.status);
  expect(onNode4).toBe(401);
  await w.rpc('cle2e_session', { op: 'unban', user_id: v.id });
  w.close();
});

test('live notifications and runtime stream calls reach sessions on other nodes', async () => {
  const target = await player(2, 'nt');
  const sender = await player(4, 'ns');
  const since = Date.now();
  const r = await sender.rpc('cle2e_notify', { user_id: target.id });
  expect(r.node).toBe('node4');
  const got = await target.wait('onnotification', (n) => n.code === 4100, 5000, since);
  expect(got).not.toBeNull();
  test.info().annotations.push({ type: 'notify-ms', description: String(got!.t - since) });

  const ch = await target.socket.joinChat(`self${runTag}`, 1, false, false);
  const sid = ch.self.session_id;
  const label = `custom${runTag}`;
  const joined = await sender.rpc('cle2e_stream', { op: 'join', label, user_id: target.id, session_id: sid });
  expect(joined.count).toBe(1);
  await sender.rpc('cle2e_stream', { op: 'send', label, data: 'hello stream' });
  expect(await target.wait('onstreamdata', (d) => d.data === 'hello stream')).not.toBeNull();
  const kicked = await sender.rpc('cle2e_stream', { op: 'kick', label, user_id: target.id, session_id: sid, node: 'node2' });
  await expect.poll(async () => (await sender.rpc('cle2e_stream', { op: 'count', label })).count).toBe(0);
  expect(kicked.node).toBe('node4');
  const discSince = Date.now();
  await sender.rpc('cle2e_session', { op: 'disconnect', session_id: sid });
  expect(await target.wait('ondisconnect', () => true, 5000, discSince)).not.toBeNull();
  sender.close();
});

test('phase 3 suspension from the console on one node kicks the player on another node', async () => {
  const p = await player(3, 'su');
  await sleep(1100);
  const since = Date.now();
  const res = await consoleApi(1, 'POST', `/v2/console/base/players/${p.id}/restrictions`, { kind: 'suspend', endsAt: new Date(Date.now() + 3600_000).toISOString(), reason: { fa: 'آزمون خوشه', en: 'cluster test' } });
  expect(res.status, JSON.stringify(res.data)).toBeLessThan(300);
  expect(await p.wait('ondisconnect', () => true, 8000, since)).not.toBeNull();
  test.info().annotations.push({ type: 'suspend-kick-ms', description: String(Date.now() - since) });
  const onNode2 = await fetch(`http://127.0.0.1:${port(2)}/v2/rpc/cle2e_whoami`, { method: 'POST', headers: { Authorization: 'Bearer ' + p.session.token }, body: '""' }).then((r) => r.status);
  expect(onNode2).toBe(401);
  const id = res.data.id || res.data.restriction?.id;
  if (id) await consoleApi(1, 'POST', `/v2/console/base/restrictions/${id}/lift`, { note: 'e2e' });
});

test('live support chat: a player message on one node wakes the console long-poll on another node at once', async () => {
  const p = await player(2, 'sc');
  const opened = await p.rpc('base_support_chat_open', { topic: `cluster ${runTag}`, body: 'سلام پشتیبانی' });
  const chatId = opened.chat?.id;
  expect(chatId, JSON.stringify(opened)).toBeTruthy();
  const first = await consoleApi(1, 'GET', `/v2/console/base/support/chats/${chatId}`);
  expect(first.status, JSON.stringify(first.data)).toBe(200);
  const poll = consoleApi(1, 'GET', `/v2/console/base/support/chats/${chatId}?wait=15&after=${first.data.lastSeq}&stamp=${encodeURIComponent(first.data.stamp)}`);
  await sleep(1500);
  const sentAt = Date.now();
  await p.rpc('base_support_chat_send', { chatId, body: 'پیام دوم از node2' });
  const res = await poll;
  const woke = Date.now() - sentAt;
  expect(res.status).toBe(200);
  expect((res.data.messages || []).some((m: any) => m.body === 'پیام دوم از node2')).toBe(true);
  test.info().annotations.push({ type: 'chat-wake-ms', description: String(woke) });
  expect(woke).toBeLessThan(1500);
  await p.rpc('base_support_chat_close', { chatId }).catch(() => {});
});
