import { consoleApi, must } from './api';
import { sql } from './db';

const P = '/v2/console/base/leaderboardx';

export const lbx = {
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
  board: async (body: any) => (await must<any>(consoleApi('POST', P + '/boards', {
    isActive: true, sortOrder: 'desc', operator: 'incr', period: 'none', tierBasis: 'none', tiers: [], rewards: [],
    source: { type: 'client', calc: 'value', multiplier: 1 }, limits: { onExceed: 'reject' }, ...body,
  }))).board,
};

export async function removeBoards(part: string) {
  sql(`UPDATE leaderboardx_result SET status='expired' WHERE status IN ('pending','claimable') AND period_id IN (SELECT id FROM leaderboardx_period WHERE board_key LIKE '%${part}%')`);
  const list = (await must<any>(consoleApi('GET', P + '/boards'))).items || [];
  for (const x of list) if (x.board.key.includes(part)) await consoleApi('DELETE', P + '/boards/' + x.board.id);
  sql(`DELETE FROM leaderboardx_score_log WHERE board_id NOT IN (SELECT id FROM leaderboardx_board);
    DELETE FROM leaderboardx_period WHERE board_key LIKE '%${part}%';`);
}

export function errorOf(res: { data: any }): { code: string; meta?: any; message?: string } {
  const raw = res.data && res.data.message;
  try {
    const env = JSON.parse(raw);
    if (env && typeof env === 'object') return env;
  } catch {}
  return { code: raw };
}
