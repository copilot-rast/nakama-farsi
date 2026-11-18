import { consoleApi, must, type Player, rpc } from './api';
import { sql } from './db';
import { API_URL } from './env';

const P = '/v2/console/base/cups';
const HTTP_KEY = process.env.HTTP_KEY || 'defaulthttpkey';

export const cup = {
  get: <T = any>(p: string) => must<T>(consoleApi('GET', P + p)),
  put: <T = any>(p: string, b: unknown) => must<T>(consoleApi('PUT', P + p, b)),
  post: <T = any>(p: string, b: unknown) => must<T>(consoleApi('POST', P + p, b)),
  del: (p: string) => consoleApi('DELETE', P + p),
  raw: (method: string, p: string, b?: unknown) => consoleApi(method, P + p, b),
};

export function rawCupSettings() {
  return sql(`SELECT value::text FROM cup_settings WHERE id=1`).trim();
}

export function restoreCupSettings(raw: string) {
  if (raw) sql(`UPDATE cup_settings SET value='${raw.replace(/'/g, "''")}'::jsonb WHERE id=1`);
}

export async function report(body: Record<string, unknown>) {
  const res = await fetch(API_URL + '/v2/rpc/cup_report?unwrap&http_key=' + encodeURIComponent(HTTP_KEY), {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
  });
  const text = await res.text();
  return { status: res.status, data: text ? JSON.parse(text) : null };
}

export async function ready(p: Player, cupId: string) {
  const r = await rpc(p, 'cup_ready', { cupId });
  if (r.status !== 200) throw new Error('cup_ready ' + r.status + ' ' + JSON.stringify(r.data));
  return r.data;
}

export function removeCups(part: string) {
  sql(`DELETE FROM cup WHERE name::text LIKE '%${part}%'; DELETE FROM cup_template WHERE key LIKE '%${part}%';`);
}
