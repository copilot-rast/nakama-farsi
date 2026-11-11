import { Client } from "@heroiclabs/nakama-js";
const PREFIX = process.env.PREFIX; const N = Number(process.env.N || 300); const VUS = Number(process.env.VUS || 1000);
const results = {};
for (const node of [1, 2, 3, 4]) {
  const c = new Client(process.env.SKEY || "defaultkey", "127.0.0.1", String(7601 + node * 10), false, 10000);
  const s = await c.authenticateDevice(`online-probe-${node}-device-xx`, true, `onlineprobe${node}`);
  const names = [];
  for (let i = 1; i <= N; i++) names.push(`${PREFIX}${1 + Math.floor((i * VUS) / (N + 1))}`);
  let online = 0, found = 0;
  for (let i = 0; i < names.length; i += 100) {
    const r = await c.getUsers(s, [], names.slice(i, i + 100));
    for (const u of r.users || []) { found++; if (u.online) online++; }
  }
  results[`node${node}`] = { online, found };
}
console.log(JSON.stringify(results));
