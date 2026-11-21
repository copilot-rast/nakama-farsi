import { API_URL, SERVER_KEY } from './env';
import type { Player } from './api';

type Waiter = { resolve: (v: any) => void; timer: ReturnType<typeof setTimeout> };

export class Chat {
  private ws: WebSocket;
  private cid = 0;
  private waiting = new Map<string, Waiter>();
  channelId = '';
  events: any[] = [];

  private constructor(ws: WebSocket) {
    this.ws = ws;
    ws.addEventListener('message', (ev) => {
      let msg: any;
      try { msg = JSON.parse(String(ev.data)); } catch { return; }
      if (msg.cid && this.waiting.has(msg.cid)) {
        const w = this.waiting.get(msg.cid)!;
        this.waiting.delete(msg.cid);
        clearTimeout(w.timer);
        w.resolve(msg);
        return;
      }
      this.events.push(msg);
    });
  }

  static open(player: Player): Promise<Chat> {
    const url = API_URL.replace(/^http/, 'ws') + '/ws?format=json&token=' + encodeURIComponent(player.token);
    const ws = new WebSocket(url);
    return new Promise((resolve, reject) => {
      ws.addEventListener('open', () => resolve(new Chat(ws)), { once: true });
      ws.addEventListener('error', () => reject(new Error('socket did not open')), { once: true });
    });
  }

  private call(body: Record<string, unknown>): Promise<any> {
    const cid = String(++this.cid);
    return new Promise((resolve) => {
      const timer = setTimeout(() => { this.waiting.delete(cid); resolve({ error: { message: 'TIMEOUT' } }); }, 10_000);
      this.waiting.set(cid, { resolve, timer });
      this.ws.send(JSON.stringify({ cid, ...body }));
    });
  }

  async join(room: string): Promise<string> {
    const r = await this.call({ channel_join: { target: room, type: 1, persistence: true, hidden: false } });
    if (r.error) throw new Error('join failed: ' + r.error.message);
    this.channelId = r.channel.id;
    return this.channelId;
  }

  async joinDirect(userId: string): Promise<string> {
    const r = await this.call({ channel_join: { target: userId, type: 2, persistence: true, hidden: false } });
    if (r.error) throw new Error('join failed: ' + r.error.message);
    this.channelId = r.channel.id;
    return this.channelId;
  }

  async send(text: string): Promise<{ ok: true; id: string } | { ok: false; message: string }> {
    return this.sendContent({ text });
  }

  async sendContent(content: Record<string, unknown>): Promise<{ ok: true; id: string } | { ok: false; message: string }> {
    const r = await this.call({ channel_message_send: { channel_id: this.channelId, content: JSON.stringify(content) } });
    if (r.error) return { ok: false, message: r.error.message };
    return { ok: true, id: r.channel_message_ack.message_id };
  }

  async rpc<T = any>(id: string, payload: unknown = {}): Promise<{ data?: T; error?: { code?: number; message: string } }> {
    const r = await this.call({ rpc: { id, payload: JSON.stringify(payload) } });
    if (r.error) return { error: r.error };
    return { data: (r.rpc?.payload ? JSON.parse(r.rpc.payload) : {}) as T };
  }

  async joinMatch(matchId: string): Promise<any> {
    const r = await this.call({ match_join: { match_id: matchId } });
    if (r.error) throw new Error('match join failed: ' + r.error.message);
    return r.match;
  }

  sendMatch(matchId: string, opCode: number, body: unknown) {
    this.ws.send(JSON.stringify({ match_data_send: { match_id: matchId, op_code: String(opCode), data: Buffer.from(JSON.stringify(body)).toString('base64') } }));
  }

  matchData(opCode: number): any[] {
    return this.events.filter((e) => e.match_data && Number(e.match_data.op_code) === opCode).map((e) => JSON.parse(Buffer.from(e.match_data.data || '', 'base64').toString('utf8') || '{}'));
  }

  async waitMatch(opCode: number, pred: (d: any) => boolean = () => true, timeout = 10_000): Promise<any> {
    const end = Date.now() + timeout;
    while (Date.now() < end) {
      const hit = this.matchData(opCode).find(pred);
      if (hit) return hit;
      await new Promise((r) => setTimeout(r, 100));
    }
    throw new Error('no match data with op ' + opCode);
  }

  close() {
    try { this.ws.close(); } catch { return; }
  }
}

export async function history(player: Player, channelId: string, limit = 20): Promise<{ id: string; text: string; code: number }[]> {
  const r = await fetch(API_URL + '/v2/channel/' + encodeURIComponent(channelId) + '?limit=' + limit + '&forward=false', { headers: { Authorization: 'Bearer ' + player.token } });
  const body: any = await r.json();
  return (body.messages || []).map((m: any) => ({ id: m.message_id, text: JSON.parse(m.content || '{}').text || '', code: Number(m.code || 0) }));
}

export function notice(message: string): any {
  try { return JSON.parse(message); } catch { return null; }
}

export async function deviceAuth(deviceId: string): Promise<{ status: number; data: any }> {
  const r = await fetch(API_URL + '/v2/account/authenticate/device?create=false', {
    method: 'POST',
    headers: { Authorization: 'Basic ' + Buffer.from(SERVER_KEY + ':').toString('base64'), 'Content-Type': 'application/json' },
    body: JSON.stringify({ id: deviceId }),
  });
  const text = await r.text();
  let data: any = null;
  try { data = JSON.parse(text); } catch { data = text; }
  return { status: r.status, data };
}

export async function refreshSession(refreshToken: string): Promise<{ status: number; data: any }> {
  const r = await fetch(API_URL + '/v2/account/session/refresh', {
    method: 'POST',
    headers: { Authorization: 'Basic ' + Buffer.from(SERVER_KEY + ':').toString('base64'), 'Content-Type': 'application/json' },
    body: JSON.stringify({ token: refreshToken }),
  });
  const text = await r.text();
  let data: any = null;
  try { data = JSON.parse(text); } catch { data = text; }
  return { status: r.status, data };
}
