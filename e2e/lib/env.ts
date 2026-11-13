import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

function readEnvFile(file: string): Record<string, string> {
  const out: Record<string, string> = {};
  if (!fs.existsSync(file)) return out;
  for (const line of fs.readFileSync(file, 'utf8').split('\n')) {
    const m = /^\s*([A-Z0-9_]+)\s*=\s*(.*)\s*$/.exec(line);
    if (m) out[m[1]] = m[2].replace(/^["']|["']$/g, '');
  }
  return out;
}

const envFile = process.env.E2E_ENV_FILE || path.join(os.homedir(), '.config/nakama-farsi/mac-dev.env');
const file = readEnvFile(envFile);
const pick = (k: string, d = '') => process.env[k] || file[k] || d;

export const ROOT = path.resolve(__dirname, '..');
export const CONSOLE_URL = pick('CONSOLE_URL', 'http://100.106.1.79:7451').replace(/\/+$/, '');
export const API_URL = pick('API_URL', 'http://100.106.1.79:7450').replace(/\/+$/, '');
export const GAME_URL = pick('GAME_URL', API_URL + '/stargame/');
export const CONSOLE_USER = pick('CONSOLE_USER', 'admin');
export const CONSOLE_PASSWORD = pick('CONSOLE_PASSWORD');
export const SERVER_KEY = pick('SERVER_KEY', 'defaultkey');
export const DB = {
  psql: pick('PSQL', '/opt/homebrew/opt/postgresql@16/bin/psql'),
  host: pick('E2E_DB_HOST', '127.0.0.1'),
  port: pick('E2E_DB_PORT', '5481'),
  name: pick('E2E_DB_NAME', 'nakama'),
  user: pick('E2E_DB_USER', 'nakama'),
  password: pick('DB_PASSWORD'),
};
export const FIXTURES = path.join(ROOT, 'fixtures');
export const FAKE_GITHUB = pick('E2E_FAKE_GITHUB_URL', 'http://127.0.0.1:7489').replace(/\/+$/, '');
export const SERVER_LOG = pick('E2E_SERVER_LOG', path.join(ROOT, '..', 'logs', 'nakama.log'));
export const AUTH_FILE = path.join(ROOT, '.auth', 'admin.json');
export const STATE_DIR = path.join(ROOT, '.auth');
export const RUN = (process.env.E2E_RUN || Date.now().toString(36).slice(-6)).toLowerCase();

if (!CONSOLE_PASSWORD) throw new Error('CONSOLE_PASSWORD missing: set it or put it in ' + envFile);
export const CHROME = process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
