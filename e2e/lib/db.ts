import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { DB, ROOT, STATE_DIR } from './env';

function run(args: string[]): string {
  return execFileSync(DB.psql, ['-h', DB.host, '-p', DB.port, '-U', DB.user, '-d', DB.name, '-At', '-v', 'ON_ERROR_STOP=1', ...args], {
    env: { ...process.env, PGPASSWORD: DB.password },
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  });
}

export function sql(query: string): string {
  return run(['-c', query]).trim();
}

export function sqlFile(name: string): string {
  return run(['-f', path.join(ROOT, 'sql', name)]).trim();
}

export function configHashes(): Record<string, string> {
  const out: Record<string, string> = {};
  for (const line of sqlFile('snapshot.sql').split('\n')) {
    const [k, v] = line.split('|');
    if (k) out[k] = v;
  }
  return out;
}

export function leftovers(): Record<string, number> {
  const out: Record<string, number> = {};
  for (const line of sqlFile('leftovers.sql').split('\n')) {
    const [k, v] = line.split('|');
    if (k && Number(v)) out[k] = Number(v);
  }
  return out;
}

const HASH_FILE = path.join(STATE_DIR, 'hashes-before.json');

export function backupConfig() {
  fs.mkdirSync(STATE_DIR, { recursive: true });
  const hashes = configHashes();
  sqlFile('backup.sql');
  fs.writeFileSync(HASH_FILE, JSON.stringify(hashes, null, 2));
  return hashes;
}

export function savedHashes(): Record<string, string> {
  return JSON.parse(fs.readFileSync(HASH_FILE, 'utf8'));
}

export function restoreConfig(final = false) {
  sqlFile('restore.sql');
  if (final) return;
  const file = path.join(STATE_DIR, 'adjusted.json');
  const adjusted: string[] = fs.existsSync(file) ? JSON.parse(fs.readFileSync(file, 'utf8')) : [];
  if (adjusted.includes('H-setup')) sql("UPDATE base_setup SET status = 'completed'");
}

export function cleanupLeftovers() {
  sqlFile('cleanup.sql');
}

export function dropBackup() {
  sql('DROP SCHEMA IF EXISTS zz_e2e_backup CASCADE');
}
