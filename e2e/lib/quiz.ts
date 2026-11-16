import { consoleApi, must } from './api';
import { sql } from './db';

const P = '/v2/console/base/quiz';

export const qz = {
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

export function removeQuiz(part: string) {
  sql(`DELETE FROM quiz_session WHERE pack_id IN (SELECT id FROM quiz_pack WHERE key LIKE '%${part}%');
    DELETE FROM quiz_question WHERE key LIKE '%${part}%' OR category LIKE '%${part}%' OR pack_id IN (SELECT id FROM quiz_pack WHERE key LIKE '%${part}%');
    DELETE FROM quiz_pack WHERE key LIKE '%${part}%';
    DELETE FROM quiz_mission WHERE key LIKE '%${part}%' AND kind='mission';
    DELETE FROM quiz_mission WHERE key LIKE '%${part}%';
    DELETE FROM quiz_result_template WHERE key LIKE '%${part}%';
    DELETE FROM quiz_report_reason WHERE key LIKE '%${part}%';
    DELETE FROM quiz_category WHERE key LIKE '%${part}%';`);
}

export function removeQuizPlayer(userId: string) {
  sql(`DELETE FROM quiz_session WHERE user_id='${userId}'; DELETE FROM quiz_user_pack WHERE user_id='${userId}'; DELETE FROM quiz_user_question WHERE user_id='${userId}';
    DELETE FROM quiz_user_mission WHERE user_id='${userId}'; DELETE FROM quiz_user_hint WHERE user_id='${userId}'; DELETE FROM quiz_user_day WHERE user_id='${userId}';
    DELETE FROM quiz_question_report WHERE user_id='${userId}'; DELETE FROM quiz_result_delivery WHERE user_id='${userId}';`);
}
