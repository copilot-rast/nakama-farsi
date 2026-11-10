import { API_URL, CONSOLE_PASSWORD, CONSOLE_URL, CONSOLE_USER, SERVER_KEY } from './env';

export type Res<T = any> = { status: number; data: T };

async function call<T>(url: string, init: RequestInit): Promise<Res<T>> {
  const r = await fetch(url, init);
  const text = await r.text();
  let data: any = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  return { status: r.status, data };
}

const basic = 'Basic ' + Buffer.from(CONSOLE_USER + ':' + CONSOLE_PASSWORD).toString('base64');

export function consoleApi<T = any>(method: string, p: string, body?: unknown, auth = basic): Promise<Res<T>> {
  return call<T>(CONSOLE_URL + p, {
    method,
    headers: { Authorization: auth, 'Content-Type': 'application/json', Accept: 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export async function must<T = any>(p: Promise<Res<T>>, ok = [200, 201]): Promise<T> {
  const r = await p;
  if (!ok.includes(r.status)) throw new Error('API ' + r.status + ' ' + JSON.stringify(r.data).slice(0, 400));
  return r.data;
}

export const base = {
  get: <T = any>(p: string) => must<T>(consoleApi('GET', '/v2/console/base' + p)),
  put: <T = any>(p: string, b: unknown) => must<T>(consoleApi('PUT', '/v2/console/base' + p, b)),
  post: <T = any>(p: string, b: unknown) => must<T>(consoleApi('POST', '/v2/console/base' + p, b)),
  del: (p: string) => consoleApi('DELETE', '/v2/console/base' + p),
};

export const eco = {
  get: <T = any>(p: string) => must<T>(consoleApi('GET', '/v2/console/economy' + p)),
  put: <T = any>(p: string, b: unknown) => must<T>(consoleApi('PUT', '/v2/console/economy' + p, b)),
  post: <T = any>(p: string, b: unknown) => must<T>(consoleApi('POST', '/v2/console/economy' + p, b)),
  del: (p: string) => consoleApi('DELETE', '/v2/console/economy' + p),
};

export const pub = {
  get: <T = any>(p: string) => must<T>(consoleApi('GET', '/v2/console/publication' + p)),
  put: <T = any>(p: string, b: unknown) => must<T>(consoleApi('PUT', '/v2/console/publication' + p, b)),
  post: <T = any>(p: string, b: unknown) => must<T>(consoleApi('POST', '/v2/console/publication' + p, b)),
  del: (p: string) => consoleApi('DELETE', '/v2/console/publication' + p),
};

export type Player = { id: string; username: string; token: string; deviceId: string };

export async function devicePlayer(deviceId: string, username?: string): Promise<Player> {
  const r = await call<any>(API_URL + '/v2/account/authenticate/device?create=true' + (username ? '&username=' + encodeURIComponent(username) : ''), {
    method: 'POST',
    headers: { Authorization: 'Basic ' + Buffer.from(SERVER_KEY + ':').toString('base64'), 'Content-Type': 'application/json' },
    body: JSON.stringify({ id: deviceId }),
  });
  if (r.status !== 200) throw new Error('device auth ' + r.status + ' ' + JSON.stringify(r.data));
  const token = r.data.token as string;
  const claims = JSON.parse(Buffer.from(token.split('.')[1], 'base64url').toString('utf8'));
  return { id: claims.uid, username: claims.usn, token, deviceId };
}

export async function rpc<T = any>(player: Player, id: string, payload: unknown = {}): Promise<Res<T>> {
  return call<T>(API_URL + '/v2/rpc/' + id + '?unwrap', {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + player.token, 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

export async function updateAccount(player: Player, body: Record<string, unknown>) {
  const r = await call(API_URL + '/v2/account', {
    method: 'PUT',
    headers: { Authorization: 'Bearer ' + player.token, 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (r.status !== 200) throw new Error('account update ' + r.status + ' ' + JSON.stringify(r.data));
}

export async function deleteUser(userId: string) {
  return consoleApi('DELETE', '/v2/console/account/' + encodeURIComponent(userId));
}

export function uid(prefix = 'e2e') {
  return prefix + '_' + Date.now().toString(36) + Math.random().toString(36).slice(2, 6);
}
