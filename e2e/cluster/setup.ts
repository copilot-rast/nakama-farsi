import fs from 'node:fs';
import path from 'node:path';
import { NODES, PREFIX, ROOT, RUN_DIR, HERE, nodes, redis, sql } from './env';

export default async function setup() {
  if (redis('ping') !== 'PONG') throw new Error('Redis is not running on the cluster test port');
  for (const n of NODES) {
    try { nodes('stop', n); } catch {}
  }
  redis('eval', "local cur='0' local n=0 repeat local r=redis.call('SCAN',cur,'MATCH',ARGV[1],'COUNT',1000) cur=r[1] for _,k in ipairs(r[2]) do redis.call('DEL',k) n=n+1 end until cur=='0' return n", '0', `${PREFIX}:*`);
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
