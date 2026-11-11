import { NODES, nodes, sql } from './env';

export default async function teardown() {
  for (const n of NODES) {
    try { nodes('stop', n); } catch {}
  }
  await new Promise((r) => setTimeout(r, 2000));
  try { sql(`DELETE FROM users WHERE username LIKE 'cle2e%'`); } catch {}
}
