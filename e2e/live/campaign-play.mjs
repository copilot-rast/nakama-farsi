import { readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { join } from 'node:path';
import { Client } from '@heroiclabs/nakama-js';

function readEnv(file) {
  const out = {};
  try {
    for (const line of readFileSync(file, 'utf8').split('\n')) {
      const m = line.match(/^\s*([A-Z0-9_]+)=(.*)$/);
      if (m) out[m[1]] = m[2].replace(/^['"]|['"]$/g, '');
    }
  } catch {}
  return out;
}

const env = readEnv(process.env.E2E_ENV_FILE || join(homedir(), '.config/nakama-farsi/mac-dev.env'));
const api = new URL(process.env.API_URL || 'http://127.0.0.1:7450');
const serverKey = process.env.SERVER_KEY || env.SERVER_KEY;
const stageKey = process.env.STAGE_KEY;
if (!stageKey) throw new Error('STAGE_KEY is required');
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const steps = [];

function errorCode(e) {
  const raw = e && (e.message || e.error || JSON.stringify(e));
  const m = String(raw).match(/[A-Z][A-Z0-9_]{3,}/g);
  return m ? m.find((x) => x !== 'RPC') || m[0] : String(raw);
}

const client = new Client(serverKey, api.hostname, api.port || (api.protocol === 'https:' ? '443' : '80'), api.protocol === 'https:');
const session = await client.authenticateDevice('e2e-live-' + Date.now() + Math.random().toString(36).slice(2, 8), true);
const socket = client.createSocket(api.protocol === 'https:', false);
await socket.connect(session, true);
const call = async (id, body) => {
  const res = await socket.rpc(id, JSON.stringify(body));
  const out = res.payload ? JSON.parse(res.payload) : null;
  steps.push({ rpc: id, request: body, response: out });
  return out;
};

const list = await call('campaign_chapters', { lang: 'fa' });
let stage = null;
for (const ch of list.chapters) for (const s of ch.stages) if (s.key === stageKey) stage = { ...s, chapterKey: ch.key, chapterSettings: ch.settings };
if (!stage) throw new Error('stage not in campaign_chapters: ' + stageKey);
const start = await call('campaign_attempt_start', { stageKey });
const wait = (stage.minSeconds || 0) + 1;
await sleep(wait * 1000);
const result = await call('campaign_attempt_complete', { attemptId: start.attemptId, stageKey, success: true, elapsedSeconds: wait - 0.5, healthPercent: 10, fields: { timesSeen: 1 }, lang: 'fa' });
let replay = null;
try {
  await call('campaign_attempt_start', { stageKey });
  replay = 'STARTED';
} catch (e) {
  replay = errorCode(e);
  steps.push({ rpc: 'campaign_attempt_start', request: { stageKey }, error: replay });
}
socket.disconnect(false);
process.stdout.write(JSON.stringify({ userId: session.user_id, stage, start, result, replay, steps }) + '\n');
