import { execFileSync } from 'node:child_process';
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

const file = readEnvFile(process.env.E2E_ENV_FILE || path.join(os.homedir(), '.config/nakama-farsi/mac-dev.env'));
export const HERE = __dirname;
export const ROOT = path.resolve(HERE, '..', '..');
export const RUN_DIR = process.env.RUN || path.join(HERE, '.run');
export const SERVER_KEY = process.env.SERVER_KEY || file.SERVER_KEY || 'defaultkey';
export const REDIS_PORT = process.env.REDIS_PORT || '6495';
export const PREFIX = process.env.CL_PREFIX || 'nkce2e';
export const NODES = (process.env.CL_NODES || '1,2,3,4').split(',').map(Number);
export const REDIS_CLI = process.env.REDIS_CLI || '/opt/homebrew/bin/redis-cli';
export const PSQL = process.env.PSQL || '/opt/homebrew/opt/postgresql@16/bin/psql';
export const DB_NAME = process.env.CL_DB || 'nakama_cluster';
export const port = (n: number) => 7601 + n * 10;

export function nodes(cmd: 'start' | 'stop' | 'kill' | 'status', ...ids: number[]) {
  return execFileSync(path.join(HERE, 'nodes.sh'), [cmd, ...ids.map(String)], {
    env: { ...process.env, RUN: RUN_DIR, CL_PREFIX: PREFIX, MODS: path.join(RUN_DIR, 'modules'), REDIS_PORT },
    encoding: 'utf8',
  });
}

export function redis(...args: string[]) {
  return execFileSync(REDIS_CLI, ['-p', REDIS_PORT, ...args], { encoding: 'utf8' }).trim();
}

export function sql(query: string) {
  const a = (process.env.NAKAMA_DATABASE_ADDRESS || file.NAKAMA_DATABASE_ADDRESS || '').replace(/^postgres(ql)?:\/\//, '');
  const up = a.split('@')[0];
  const user = up.split(':')[0];
  const password = up.slice(user.length + 1);
  return execFileSync(PSQL, ['-h', '127.0.0.1', '-p', '5481', '-U', user, '-d', DB_NAME, '-Atc', query], { env: { ...process.env, PGPASSWORD: password }, encoding: 'utf8' }).trim();
}

export function nodePid(n: number) {
  return Number(fs.readFileSync(path.join(RUN_DIR, `n${n}`, 'pid'), 'utf8').trim());
}

export const CONSOLE_USER = process.env.CONSOLE_USER || file.CONSOLE_USER || 'admin';
export const CONSOLE_PASSWORD = process.env.CONSOLE_PASSWORD || file.CONSOLE_PASSWORD || '';
export const consolePort = (n: number) => 7602 + n * 10;

export async function consoleApi(node: number, method: string, p: string, body?: unknown) {
  const r = await fetch(`http://127.0.0.1:${consolePort(node)}${p}`, {
    method,
    headers: { Authorization: 'Basic ' + Buffer.from(CONSOLE_USER + ':' + CONSOLE_PASSWORD).toString('base64'), 'Content-Type': 'application/json', Accept: 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await r.text();
  let data: any = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  return { status: r.status, data };
}
