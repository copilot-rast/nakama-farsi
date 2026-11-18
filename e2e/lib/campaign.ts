import { consoleApi, must } from './api';
import { sql } from './db';

const P = '/v2/console/base/campaign';

export const cg = {
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

export function removeCampaign(part: string) {
  sql(`DELETE FROM campaign_attempt WHERE stage_id IN (SELECT id FROM campaign_stage WHERE key LIKE '%${part}%');
    DELETE FROM campaign_user_stage WHERE stage_id IN (SELECT id FROM campaign_stage WHERE key LIKE '%${part}%');
    DELETE FROM campaign_stage WHERE key LIKE '%${part}%';
    DELETE FROM campaign_user_chapter WHERE chapter_id IN (SELECT id FROM campaign_chapter WHERE key LIKE '%${part}%');
    DELETE FROM campaign_chapter WHERE key LIKE '%${part}%';
    DELETE FROM campaign_claim WHERE freeride_id IN (SELECT id FROM campaign_freeride WHERE key LIKE '%${part}%');
    DELETE FROM campaign_freeride WHERE key LIKE '%${part}%';
    DELETE FROM campaign_reward_group WHERE group_key LIKE '%${part}%';`);
}
