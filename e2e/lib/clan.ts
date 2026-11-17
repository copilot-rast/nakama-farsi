import { consoleApi, must, rpc, type Player } from './api';
import { sql } from './db';

function client(prefix: string) {
  const P = '/v2/console/base/' + prefix;
  return {
    get: <T = any>(p: string) => must<T>(consoleApi('GET', P + p)),
    put: <T = any>(p: string, b: unknown) => must<T>(consoleApi('PUT', P + p, b)),
    post: <T = any>(p: string, b: unknown) => must<T>(consoleApi('POST', P + p, b)),
    del: (p: string) => consoleApi('DELETE', P + p),
    raw: (method: string, p: string, b?: unknown) => consoleApi(method, P + p, b),
    settings: async () => (await must<any>(consoleApi('GET', P + '/settings'))).settings,
    setSettings: async (change: (s: any) => void) => {
      const s = (await must<any>(consoleApi('GET', P + '/settings'))).settings;
      change(s);
      return must<any>(consoleApi('PUT', P + '/settings', s));
    },
  };
}

export const tro = client('trophies');
export const cln = client('clans');

export function rawSettings(table: 'trophy_settings' | 'clan_settings') {
  return sql(`SELECT value::text FROM ${table} WHERE id=1`).trim();
}

export function restoreSettings(table: 'trophy_settings' | 'clan_settings', raw: string) {
  if (raw) sql(`UPDATE ${table} SET value='${raw.replace(/'/g, "''")}'::jsonb WHERE id=1`);
}

export async function createClan(owner: Player, name: string, extra: Record<string, unknown> = {}) {
  const r = await rpc(owner, 'clan_create', { name, joinType: 'open', ...extra });
  if (r.status !== 200) throw new Error('clan_create ' + r.status + ' ' + JSON.stringify(r.data));
  return r.data.clan || r.data;
}

export async function joinClan(player: Player, clanId: string) {
  const r = await rpc(player, 'clan_join', { clanId });
  if (r.status !== 200) throw new Error('clan_join ' + r.status + ' ' + JSON.stringify(r.data));
  return r.data;
}

export function removeTrophyPart(part: string) {
  sql(`DELETE FROM trophy_league WHERE key LIKE '%${part}%';`);
}

export function removeTrophyPlayer(userId: string) {
  sql(`DELETE FROM trophy_prize WHERE user_id='${userId}'; DELETE FROM trophy_log WHERE user_id='${userId}'; DELETE FROM trophy_player WHERE user_id='${userId}';`);
}

export function removeWarPart(part: string) {
  sql(`DELETE FROM clan_war_round WHERE config_key LIKE '%${part}%'; DELETE FROM clan_war_config WHERE key LIKE '%${part}%';`);
}

export async function removeClan(clanId: string) {
  if (!clanId) return;
  await cln.del('/list/' + clanId);
  sql(`DELETE FROM clan WHERE group_id='${clanId}'; DELETE FROM clan_member WHERE group_id='${clanId}';`);
}
