import { Client } from "@heroiclabs/nakama-js";

const DUR = Number(process.env.DUR || 30);
const KEY = process.env.SKEY || "defaultkey";
const port = (n) => 7601 + n * 10;
const tag = Date.now().toString(36);

async function player(node, name) {
  const client = new Client(KEY, "127.0.0.1", String(port(node)), false, 10000);
  const session = await client.authenticateDevice(`probe-${tag}-${name}-device`, true, `probe${tag}${name}`.slice(0, 40));
  const socket = client.createSocket(false, false);
  await socket.connect(session, false);
  return { client, session, socket };
}

function stats(arr) {
  if (!arr.length) return { n: 0 };
  const s = [...arr].sort((a, b) => a - b);
  const q = (x) => Number(s[Math.min(s.length - 1, Math.floor(x * s.length))].toFixed(2));
  return { n: s.length, p50: q(0.5), p95: q(0.95), p99: q(0.99), max: Number(s[s.length - 1].toFixed(2)) };
}

async function pair(from, to) {
  const a = await player(from, `a${from}${to}`);
  const b = await player(to, `b${from}${to}`);
  const m = await a.socket.createMatch();
  await b.socket.joinMatch(m.match_id);
  await new Promise((r) => setTimeout(r, 300));
  const lat = [];
  b.socket.onmatchdata = (d) => {
    const t = Number(new TextDecoder().decode(d.data));
    lat.push(performance.now() - t);
  };
  let sent = 0;
  const end = Date.now() + DUR * 1000;
  while (Date.now() < end) {
    await a.socket.sendMatchState(m.match_id, 1, String(performance.now()));
    sent++;
    await new Promise((r) => setTimeout(r, 100));
  }
  await new Promise((r) => setTimeout(r, 1000));
  a.socket.disconnect(false);
  b.socket.disconnect(false);
  return { from, to, sent, lost: sent - lat.length, ...stats(lat) };
}

const [cross, same] = await Promise.all([pair(2, 3), pair(2, 2)]);
console.log(JSON.stringify({ cross, same }));
process.exit(0);
