import fs from 'node:fs';
import path from 'node:path';
import { NODES, PREFIX, ROOT, RUN_DIR, HERE, nodes, redis, sql } from './env';

export default async function setup() {
  if (redis('ping') !== 'PONG') throw new Error('Redis is not running on the cluster test port');
  for (const n of NODES) {
    try { nodes('stop', n); } catch {}
  }
  const keys = redis('--scan', '--pattern', `${PREFIX}:*`).split('\n').filter(Boolean);
  for (let i = 0; i < keys.length; i += 500) redis('del', ...keys.slice(i, i + 500));
  const mods = path.join(RUN_DIR, 'modules');
  fs.rmSync(mods, { recursive: true, force: true });
  fs.mkdirSync(mods, { recursive: true });
  const built = path.join(ROOT, '.local', 'cluster', 'modules');
  for (const f of fs.readdirSync(built)) fs.symlinkSync(path.join(built, f), path.join(mods, f));
  for (const f of fs.readdirSync(path.join(HERE, 'modules'))) fs.copyFileSync(path.join(HERE, 'modules', f), path.join(mods, f));
  try { sql(`DELETE FROM users WHERE username LIKE 'cle2e%'`); } catch {}
  nodes('start', ...NODES);
  await new Promise((r) => setTimeout(r, 3000));
}
