import crypto from 'node:crypto';
import http from 'node:http';
import type { AddressInfo } from 'node:net';
import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, openSettings, root, settle, toasts } from '../../lib/console';

const k = uid();
const APP_ID = crypto.randomUUID();
const ISSUER = 'nakama-farsi-' + k.slice(-4);
const KID = 'game-' + k.slice(-6);
const SECRET = 'comms-hmac-' + k + '-0123456789abcdef';
const CK = 'ck_e2e_' + k;
const users: string[] = [];
let P: Player;
let fake: http.Server;
let fakeUrl = '';
const sessions: { sub: string; name: string; life: number }[] = [];
const calls: string[] = [];

test.describe.configure({ mode: 'serial' });

function jwt(token: string) {
  const [head, body] = token.split('.').slice(0, 2).map((p) => JSON.parse(Buffer.from(p, 'base64url').toString('utf8')));
  return { head, body };
}

function verify(token: string): { sub: string; name: string; life: number } | null {
  const parts = token.split('.');
  if (parts.length !== 3) return null;
  const mac = crypto.createHmac('sha256', SECRET).update(parts[0] + '.' + parts[1]).digest('base64url');
  if (mac !== parts[2]) return null;
  const { head, body } = jwt(token);
  const now = Math.floor(Date.now() / 1000);
  if (head.alg !== 'HS256' || head.kid !== KID) return null;
  if (body.aud !== APP_ID || body.iss !== ISSUER || !/^[A-Za-z0-9._~@+-]{1,128}$/.test(body.sub)) return null;
  if (typeof body.iat !== 'number' || body.iat > now + 30 || body.exp < now - 30 || body.exp - body.iat > 600) return null;
  return { sub: body.sub, name: body.name || '', life: body.exp - body.iat };
}

function startFake(): Promise<void> {
  fake = http.createServer((req, res) => {
    calls.push(req.method + ' ' + req.url);
    const auth = (req.headers.authorization || '').replace(/^Bearer /, '');
    const send = (status: number, body?: unknown) => { res.writeHead(status, { 'Content-Type': 'application/json' }); res.end(body === undefined ? '' : JSON.stringify(body)); };
    if (req.url === '/healthz') return send(200, { status: 'ok' });
    if (req.method === 'POST' && req.url === '/v1/sessions') {
      const ok = verify(auth);
      if (!ok) return send(401, { error: { code: 'unauthorized', message: 'unauthorized' } });
      sessions.push(ok);
      return send(201, { session_token: 'sess-' + ok.sub, session_expires_at: new Date(Date.now() + 900_000).toISOString(), resume_token: 'cr_x', user: { id: ok.sub, display_name: ok.name } });
    }
    if (req.method === 'DELETE' && req.url === '/v1/sessions/current') return send(204);
    if (auth !== CK) return send(401, { error: { code: 'unauthorized' } });
    if (req.method === 'GET' && req.url?.startsWith('/v1/server/game-rooms/')) return send(404, { error: { code: 'not_found' } });
    send(404, { error: { code: 'not_found' } });
  });
  return new Promise((resolve) => fake.listen(0, '127.0.0.1', () => { fakeUrl = 'http://127.0.0.1:' + (fake.address() as AddressInfo).port; resolve(); }));
}

test.beforeAll(async () => {
  await startFake();
  P = await devicePlayer('e2e-cut-' + k, ('cut' + k).replace(/_/g, '').slice(0, 24));
  users.push(P.id);
  await updateAccount(P, { lang_tag: 'fa', display_name: 'گفتگو ' + k.slice(-4) });
});

test.afterAll(async () => {
  fake?.close();
  sql(`UPDATE base_voice_settings SET enabled=false, provider='', config='{}', secrets_sealed='', ttl_seconds=600, open_rooms=false, can_publish=true, user_tokens=false, user_token_ttl_seconds=300, user_tokens_per_minute=10 WHERE id=1; DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%';`);
  for (const id of users) await deleteUser(id);
});

test.describe('comms user tokens', () => {
  test('console fields, secret length check, connection test exchanges a user token, sealed secret, voice stays off', async ({ page }) => {
    await openSettings(page, 'voice');
    const form = page.locator('[data-settings-form="voice"]');
    await form.locator('select').first().selectOption('comms');
    await expect(form).toContainText('توکن کاربر comms (چت، تماس و جلسه)');
    await expect(form).toContainText('POST /v1/sessions');
    await expect(form).toContainText('عمر توکن کاربر (ثانیه)');
    await expect(field(form, 'عمر توکن کاربر (ثانیه)').locator('input')).toHaveValue(/^(300|۳۰۰)$/);
    await field(form, 'نشانی سرویس comms برای سرور').locator('input').fill(fakeUrl);
    await field(form, 'نشانی عمومی comms').locator('input').fill('https://comms.example.com');
    await form.locator('[data-voice-secret="comms.serverKey"]').fill(CK);
    await form.getByText('توکن کاربر comms روشن باشد').click();
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'شناسهٔ برنامه (App ID)')).toContainText('این را پر کنید');

    await field(form, 'شناسهٔ برنامه (App ID)').locator('input').fill('not-a-uuid');
    await field(form, 'صادرکنندهٔ توکن (issuer)').locator('input').fill(ISSUER);
    await field(form, 'شناسهٔ کلید توکن (kid)').locator('input').fill(KID);
    await form.locator('[data-voice-secret="comms.tokenSecret"]').fill('short-secret');
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'شناسهٔ برنامه (App ID)')).toContainText('UUID');
    await field(form, 'شناسهٔ برنامه (App ID)').locator('input').fill(APP_ID);
    await form.locator('[data-settings-save]').click();
    await expect(field(form, 'رمز توکن')).toContainText('دست‌کم ۳۲ بایت');

    await form.locator('[data-voice-secret="comms.tokenSecret"]').fill('wrong-secret-wrong-secret-wrong-secret');
    await form.locator('[data-voice-test]').click();
    await expect(form.locator('[data-voice-result]')).toContainText('comms توکن کاربر را نپذیرفت');
    await form.locator('[data-voice-secret="comms.tokenSecret"]').fill(SECRET);
    await form.locator('[data-voice-test]').click();
    await expect(form.locator('[data-voice-result]')).toContainText('توکن کاربر را هم comms پذیرفت');
    expect(sessions.map((s) => s.sub)).toContain('nakama-connection-test');
    expect(calls).toContain('DELETE /v1/sessions/current');

    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    await settle(page);
    await expect(form.locator('[data-voice-state]')).toContainText('چت صوتی خاموش است');
    await expect(form.locator('[data-voice-ut-state]')).toContainText('توکن کاربر comms روشن است');
    await expect(form.locator('[data-voice-secret="comms.tokenSecret"]')).toHaveValue('********');
    const saved = await base.get('/voice/settings');
    expect(saved).toMatchObject({ enabled: false, provider: 'comms', userTokens: true, userTokenTtlSeconds: 300, userTokensPerMinute: 10, userTokenReady: true,
      config: { appId: APP_ID, tokenIssuer: ISSUER, tokenKeyId: KID }, secrets: { tokenSecret: '********', serverKey: '********' } });
    expect(sql(`SELECT position('${SECRET}' in secrets_sealed) + position('${SECRET}' in config::text) FROM base_voice_settings`)).toBe('0');
    expect(sql(`SELECT count(*) FROM console_audit_log WHERE metadata::text LIKE '%${SECRET}%'`)).toBe('0');
  });

  test('comms_user_token signs for the player, comms accepts it, the switch and the provider turn it off', async () => {
    const res = await rpc(P, 'comms_user_token', {});
    expect(res.status).toBe(200);
    expect(res.data).toMatchObject({ enabled: true, url: 'https://comms.example.com', appId: APP_ID });
    const { head, body } = jwt(res.data.token);
    expect(head).toEqual({ alg: 'HS256', typ: 'JWT', kid: KID });
    expect(body).toMatchObject({ aud: APP_ID, iss: ISSUER, sub: P.id, name: 'گفتگو ' + k.slice(-4) });
    expect(body.exp - body.iat).toBe(300);
    expect(Date.parse(res.data.expiresAt) / 1000).toBe(body.exp);

    const ex = await fetch(fakeUrl + '/v1/sessions', { method: 'POST', headers: { Authorization: 'Bearer ' + res.data.token } });
    expect(ex.status).toBe(201);
    expect(sessions.at(-1)).toEqual({ sub: P.id, name: 'گفتگو ' + k.slice(-4), life: 300 });

    const s = await base.get('/voice/settings');
    await base.put('/voice/settings', { enabled: false, provider: 'comms', ttlSeconds: 600, userTokens: true, userTokenTtlSeconds: 600, userTokensPerMinute: 2, config: s.config, secrets: { serverKey: '********', tokenSecret: '********' } });
    sql(`DELETE FROM cluster_rate_limit WHERE key LIKE 'voice-%'`);
    const long = (await rpc(P, 'comms_user_token', {})).data;
    expect(jwt(long.token).body.exp - jwt(long.token).body.iat).toBe(600);
    await rpc(P, 'comms_user_token', {});
    const limited = await rpc(P, 'comms_user_token', {});
    expect(JSON.stringify(limited.data)).toContain('COMMS_TOKEN_RATE_LIMITED');
    expect(limited.status).toBe(429);

    await base.put('/voice/settings', { enabled: false, provider: 'comms', ttlSeconds: 600, userTokens: false, config: s.config, secrets: { serverKey: '********', tokenSecret: '********' } });
    expect((await rpc(P, 'comms_user_token', {})).data).toEqual({ enabled: false, reason: 'COMMS_TOKENS_DISABLED' });
    sql(`UPDATE base_voice_settings SET user_tokens=true, provider='livekit' WHERE id=1`);
    expect((await rpc(P, 'comms_user_token', {})).data).toEqual({ enabled: false, reason: 'COMMS_NOT_CONFIGURED' });
  });
});
