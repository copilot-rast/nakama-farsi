import crypto from 'node:crypto';
import http from 'node:http';
import type { AddressInfo } from 'node:net';
import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, open, openSettings, root, rows, savePanel, settle, toasts } from '../../lib/console';

const k = uid();
const gameKey = 'e2e_' + k.replace(/^e2e_/, '') + '_voice';
const LK_KEY = 'APIe2e' + k.slice(-4);
const LK_SECRET = 'lk-secret-' + k + '-0123456789';
const CK = 'ck_e2e_' + k;
const users: string[] = [];
let P: Player;
let gameId = '';
let fake: http.Server;
let fakeUrl = '';
const calls: string[] = [];

test.describe.configure({ mode: 'serial' });

function jwt(token: string) {
  const [head, body] = token.split('.').slice(0, 2).map((p) => JSON.parse(Buffer.from(p, 'base64url').toString('utf8')));
  return { head, body };
}

function startFake(): Promise<void> {
  fake = http.createServer((req, res) => {
    calls.push(req.method + ' ' + req.url);
    const auth = req.headers.authorization || '';
    const send = (status: number, body: unknown) => { res.writeHead(status, { 'Content-Type': 'application/json' }); res.end(JSON.stringify(body)); };
    if (req.url === '/healthz') return send(200, { status: 'ok' });
    if (req.url === '/twirp/livekit.RoomService/ListRooms') {
      const token = auth.replace(/^Bearer /, '');
      try {
        const [h, b, s] = token.split('.');
        const mac = crypto.createHmac('sha256', LK_SECRET).update(h + '.' + b).digest('base64url');
        return mac === s ? send(200, { rooms: [] }) : send(401, { msg: 'invalid token' });
      } catch { return send(401, {}); }
    }
    if (auth !== 'Bearer ' + CK) return send(401, { error: { code: 'unauthorized', message: 'bad key' } });
    if (req.method === 'GET' && req.url?.startsWith('/v1/server/game-rooms/')) return send(404, { error: { code: 'not_found' } });
    if (req.method === 'POST' && req.url === '/v1/server/game-rooms') return send(201, {});
    const m = /^\/v1\/server\/game-rooms\/([^/]+)\/players\/([^/]+)\/token$/.exec(req.url || '');
    if (req.method === 'POST' && m) return send(200, { token: 'comms-join-' + decodeURIComponent(m[1]) + '-' + m[2], expires_at: new Date(Date.now() + 600_000).toISOString() });
    send(404, { error: { code: 'not_found' } });
  });
  return new Promise((resolve) => fake.listen(0, '127.0.0.1', () => { fakeUrl = 'http://127.0.0.1:' + (fake.address() as AddressInfo).port; resolve(); }));
}

test.beforeAll(async () => {
  await startFake();
  P = await devicePlayer('e2e-vc-' + k, ('vc' + k).replace(/_/g, '').slice(0, 24));
  users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa', display_name: 'صدا ' + k.slice(-4) });
  gameId = (await eco.post('/games', { key: gameKey, name: { fa: 'بازی ویس ' + k.slice(-4), en: 'Voice game' }, isActive: true })).id;
});

test.afterAll(async () => {
  fake?.close();
  sql(`UPDATE base_voice_settings SET enabled=false, provider='', config='{}', secrets_sealed='', ttl_seconds=600, open_rooms=false, can_publish=true WHERE id=1; DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%';`);
  if (gameId) await eco.del('/games/' + gameId);
  for (const id of users) await deleteUser(id);
});

test.describe('voice: server settings', () => {
  test('provider form, test connection with a wrong and a right secret, sealed masked secret, audit log', async ({ page }) => {
    await openSettings(page, 'voice');
    const form = root(page).locator('[data-settings-form="voice"]');
    await expect(form.locator('[data-voice-state]')).toContainText('چت صوتی خاموش است');
    await form.locator('select').first().selectOption('livekit');
    await expect(form).toContainText('توکن لایوکیت همین‌جا با کلید و رمز API امضا می‌شود');
    await expect(form).toContainText('فقط این برچسب‌ها ویس بگیرند');
    await expect(form).toContainText('بیشترین بازیکن در اتاق ویس');
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'نشانی سرور لایوکیت')).toContainText('این فیلد لازم است');
    await field(form, 'نشانی سرور لایوکیت').locator('input').fill(fakeUrl.replace('http://', 'ws://'));
    await field(form, 'کلید API').locator('input').fill(LK_KEY);
    await form.locator('[data-voice-secret="livekit.apiSecret"]').fill('wrong-secret');
    await form.locator('[data-voice-test]').click();
    await expect(form.locator('[data-voice-result]')).toContainText('کلید یا رمز را سرویس ویس نپذیرفت');
    await form.locator('[data-voice-secret="livekit.apiSecret"]').fill(LK_SECRET);
    await form.locator('[data-voice-test]').click();
    await expect(form.locator('[data-voice-result]')).toContainText('اتصال به لایوکیت (LiveKit) درست است');
    await form.getByText('چت صوتی روشن باشد').click();
    await form.getByText('اتاق‌های باز هر بازی').click();
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    await settle(page);
    await expect(form.locator('[data-voice-state]')).toContainText('چت صوتی روشن است با لایوکیت');
    await expect(form.locator('[data-voice-secret="livekit.apiSecret"]')).toHaveValue('********');
    const saved = await base.get('/voice/settings');
    expect(saved).toMatchObject({ enabled: true, provider: 'livekit', openRooms: true, canPublish: true, ttlSeconds: 600, secrets: { apiSecret: '********' }, config: { apiKey: LK_KEY } });
    expect(sql(`SELECT position('${LK_SECRET}' in secrets_sealed) FROM base_voice_settings`)).toBe('0');
    expect(sql(`SELECT count(*) FROM console_audit_log WHERE metadata::text LIKE '%${LK_SECRET}%'`)).toBe('0');
    expect(Number(sql(`SELECT count(*) FROM console_audit_log WHERE message LIKE 'base %voice%' AND create_time > now() - interval '5 minutes'`))).toBeGreaterThanOrEqual(3);
    await form.locator('[data-voice-test]').click();
    await expect(form.locator('[data-voice-result]')).toContainText('درست است');
  });
});

test.describe('voice: tokens', () => {
  test('voice_token for an open room signs a LiveKit token, refuses strangers, follows the game switch, switches to comms and turns off', async ({ page }) => {
    const res = await rpc(P, 'voice_token', { room: 'lobby', gameKey });
    expect(res.status).toBe(200);
    expect(res.data).toMatchObject({ enabled: true, provider: 'livekit', url: fakeUrl.replace('http://', 'ws://'), room: 'room.' + gameKey + '.lobby', identity: P.id, canPublish: true, gameKey });
    const { head, body } = jwt(res.data.token);
    expect(head).toMatchObject({ alg: 'HS256', typ: 'JWT' });
    expect(body).toMatchObject({ iss: LK_KEY, sub: P.id, name: 'صدا ' + k.slice(-4), video: { roomJoin: true, room: 'room.' + gameKey + '.lobby', canPublish: true, canSubscribe: true } });
    expect(body.exp - body.iat).toBe(600);
    expect(Date.parse(res.data.expiresAt)).toBeGreaterThan(Date.now() + 500_000);

    const stranger = await rpc(P, 'voice_token', { matchId: '00000000-0000-4000-8000-000000000001.nkfarsi1' });
    expect(JSON.stringify(stranger.data)).toContain('MATCH_NOT_FOUND');
    const two = await rpc(P, 'voice_token', { matchId: 'x', room: 'lobby' });
    expect(JSON.stringify(two.data)).toContain('VOICE_ROOM_REQUIRED');

    await open(page, '#/economy/games/' + gameId);
    const panel = page.locator('[data-economy="panel"]');
    await expect(panel.locator('form')).toBeVisible();
    await panel.getByText('چت صوتی', { exact: true }).click();
    await savePanel(page);
    await open(page, '#/economy/games?q=' + gameKey);
    await expect(rows(page).filter({ hasText: gameKey })).toContainText('بی‌ویس');
    expect((await rpc(P, 'voice_token', { room: 'lobby', gameKey })).data).toEqual({ enabled: false, reason: 'GAME_VOICE_DISABLED', gameKey });
    expect((await eco.get('/games/' + gameId)).voiceEnabled).toBe(false);
    sql(`UPDATE economy_game SET voice_enabled=true WHERE id='${gameId}'`);

    const quiet = await base.post('/tags', { key: 'e2e_' + k + '_vquiet', keyType: 'user', name: { fa: 'فقط شنونده ' + k.slice(-4) }, description: {}, isPublic: false, metadata: {}, sort: 0 });
    try {
      await base.put('/players/' + P.id + '/tags', { tagIds: [quiet.id] });
      const lk = await base.get('/voice/settings');
      await base.put('/voice/settings', { enabled: true, provider: 'livekit', ttlSeconds: 600, openRooms: true, canPublish: true, config: lk.config, secrets: { apiSecret: '********' }, listenOnlyTagIds: [quiet.id], tokensPerMinute: 30, maxRoomPlayers: 0 });
      const listen = (await rpc(P, 'voice_token', { room: 'lobby', gameKey })).data;
      expect(listen.canPublish).toBe(false);
      expect(jwt(listen.token).body.video.canPublish).toBe(false);
      await base.put('/voice/settings', { enabled: true, provider: 'livekit', ttlSeconds: 600, openRooms: true, canPublish: true, config: lk.config, secrets: { apiSecret: '********' }, allowedTagIds: [quiet.id], tokensPerMinute: 30 });
      await base.put('/players/' + P.id + '/tags', { tagIds: [] });
      expect((await rpc(P, 'voice_token', { room: 'lobby', gameKey })).data).toEqual({ enabled: false, reason: 'VOICE_NOT_FOR_PLAYER' });
    } finally {
      await base.put('/players/' + P.id + '/tags', { tagIds: [] });
      await base.put('/voice/settings', { enabled: false, ttlSeconds: 600 });
      await base.del('/tags/' + quiet.id);
    }

    await base.put('/voice/settings', { enabled: true, provider: 'comms', ttlSeconds: 900, openRooms: true, canPublish: true, config: { url: fakeUrl, publicUrl: 'https://comms.example.com' }, secrets: { serverKey: CK } });
    const comms = await rpc(P, 'voice_token', { room: 'lobby', gameKey });
    expect(comms.data).toMatchObject({ enabled: true, provider: 'comms', url: 'https://comms.example.com', token: 'comms-join-room.' + gameKey + '.lobby-' + P.id, identity: P.id });
    expect(calls).toContain('POST /v1/server/game-rooms/room.' + gameKey + '.lobby/players/' + P.id + '/token');

    await base.put('/voice/settings', { enabled: false, provider: 'comms', ttlSeconds: 900, openRooms: true, canPublish: true, config: { url: fakeUrl }, secrets: { serverKey: '********' } });
    expect((await rpc(P, 'voice_token', { room: 'lobby', gameKey })).data).toEqual({ enabled: false, reason: 'VOICE_DISABLED' });
  });
});
