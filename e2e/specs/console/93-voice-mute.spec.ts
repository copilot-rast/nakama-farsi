import http from 'node:http';
import type { AddressInfo } from 'node:net';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { noHorizontalScroll, open, root, toasts } from '../../lib/console';

const k = uid();
const gameKey = 'e2e_' + k.replace(/^e2e_/, '') + '_vm';
const LK_KEY = 'APIvm' + k.slice(-4);
const LK_SECRET = 'lk-secret-vm-' + k + '-0123456789';
const users: string[] = [];
const calls: { path: string; body: any }[] = [];
let A: Player;
let B: Player;
let gameId = '';
let room = '';
let fake: http.Server;
let fakeUrl = '';

test.describe.configure({ mode: 'serial' });

function claims(token: string) {
  return JSON.parse(Buffer.from(token.split('.')[1], 'base64url').toString('utf8'));
}

function startFake(): Promise<void> {
  fake = http.createServer((req, res) => {
    let raw = '';
    req.on('data', (c) => { raw += c; });
    req.on('end', () => {
      let body: any = {};
      try { body = JSON.parse(raw || '{}'); } catch { body = {}; }
      calls.push({ path: req.url || '', body });
      res.writeHead(200, { 'Content-Type': 'application/json' });
      if (req.url === '/twirp/livekit.RoomService/ListParticipants') return res.end(JSON.stringify({ participants: [{ identity: A.id, name: 'A', permission: { canPublish: true } }] }));
      res.end('{}');
    });
  });
  return new Promise((resolve) => fake.listen(0, '127.0.0.1', () => { fakeUrl = 'http://127.0.0.1:' + (fake.address() as AddressInfo).port; resolve(); }));
}

test.beforeAll(async () => {
  await startFake();
  A = await devicePlayer('e2e-vm-a-' + k, ('vma' + k).replace(/_/g, '').slice(0, 24));
  B = await devicePlayer('e2e-vm-b-' + k, ('vmb' + k).replace(/_/g, '').slice(0, 24));
  users.push(A.id, B.id);
  await updateAccount(A, { display_name: 'گوینده ' + k.slice(-4) });
  gameId = (await eco.post('/games', { key: gameKey, name: { fa: 'بازی بی‌صدا ' + k.slice(-4), en: 'Mute game' }, isActive: true })).id;
  await base.put('/voice/settings', { enabled: true, provider: 'livekit', ttlSeconds: 600, canPublish: true, openRooms: true, tokensPerMinute: 30,
    config: { url: fakeUrl.replace('http://', 'ws://'), apiKey: LK_KEY }, secrets: { apiSecret: LK_SECRET } });
});

test.afterAll(async () => {
  fake?.close();
  sql(`UPDATE base_voice_settings SET enabled=false, provider='', config='{}', secrets_sealed='', ttl_seconds=600, open_rooms=false, can_publish=true WHERE id=1; DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%';`);
  if (room) sql(`DELETE FROM base_voice_mute WHERE room='${room}'; DELETE FROM base_voice_room WHERE room='${room}';`);
  if (gameId) await eco.del('/games/' + gameId);
  for (const id of users) await deleteUser(id);
});

test.describe('voice: staff mute', () => {
  test('live rooms list both players; mute revokes publish on LiveKit and in the next token; unmute gives it back', async ({ page }) => {
    const first = await rpc(A, 'voice_token', { room: 'lobby', gameKey });
    expect(first.status).toBe(200);
    expect(first.data).toMatchObject({ enabled: true, canPublish: true, staffMuted: false });
    room = first.data.room;
    expect((await rpc(B, 'voice_token', { room: 'lobby', gameKey })).data.room).toBe(room);

    await open(page, '#/moderation/voice-rooms?q=' + gameKey);
    const rowA = root(page).locator('tr', { has: page.locator(`[data-vr-action="${A.id}"]`) });
    const rowB = root(page).locator('tr', { has: page.locator(`[data-vr-action="${B.id}"]`) });
    await expect(rowA.locator(`[data-vr-room="${room}"]`)).toContainText('اتاق باز');
    await expect(rowA).toContainText('گوینده');
    await expect(rowA).toContainText('وصل');
    await expect(rowB).toContainText('وصل نیست');
    await expect(rowA.locator('[data-vr-mic="open"]')).toContainText('می‌تواند صحبت کند');

    await rowA.locator(`[data-vr-action="${A.id}"]`).click();
    await expect(page.locator('[data-economy="dialog"]')).toContainText('بی‌صدا شود؟');
    await page.locator('[data-economy="dialog"] button', { hasText: 'بی‌صدا کردن' }).click();
    await expect(toasts(page).last()).toContainText('بازیکن بی‌صدا شد');
    await expect(rowA.locator('[data-vr-mic="muted"]')).toContainText('بی‌صدا به‌دست پشتیبانی');
    const update = calls.filter((c) => c.path === '/twirp/livekit.RoomService/UpdateParticipant').pop()!;
    expect(update.body).toMatchObject({ room, identity: A.id, permission: { canPublish: false, canSubscribe: true } });

    const muted = await rpc(A, 'voice_token', { room: 'lobby', gameKey });
    expect(muted.data).toMatchObject({ canPublish: false, staffMuted: true });
    expect(claims(muted.data.token).video.canPublish).toBe(false);
    expect((await rpc(B, 'voice_token', { room: 'lobby', gameKey })).data.canPublish).toBe(true);
    expect(sql(`SELECT muted_by <> '' FROM base_voice_mute WHERE room='${room}' AND user_id='${A.id}'`)).toBe('t');

    await rowA.locator(`[data-vr-action="${A.id}"]`).click();
    await expect(toasts(page).last()).toContainText('بی‌صدا برداشته شد');
    expect(calls.filter((c) => c.path === '/twirp/livekit.RoomService/UpdateParticipant').pop()!.body.permission.canPublish).toBe(true);
    expect((await rpc(A, 'voice_token', { room: 'lobby', gameKey })).data).toMatchObject({ canPublish: true, staffMuted: false });
  });

  test('a mute of a player who is not in the room is refused, and the page reads in English', async ({ page }) => {
    const res = await consoleApi('POST', '/v2/console/base/voice/rooms/' + encodeURIComponent(room) + '/mute', { userId: '00000000-0000-4000-8000-00000000abcd' });
    expect(res.status).toBe(404);
    expect(JSON.stringify(res.data)).toContain('VOICE_PARTICIPANT_NOT_FOUND');
    await open(page, '#/moderation/voice-rooms?q=' + gameKey);
    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(root(page)).toContainText('Voice rooms');
    await expect(root(page).locator(`[data-vr-action="${A.id}"]`)).toHaveText('Mute');
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  test('the voice rooms and match chat pages fit a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const hash of ['#/moderation/voice-rooms?q=' + gameKey, '#/moderation/stickers/match-chat', '#/moderation/stickers/channels']) {
      await open(page, hash);
      await test.step(hash, () => noHorizontalScroll(page));
    }
  });
});
