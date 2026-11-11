import { Client, Session, Socket } from '@heroiclabs/nakama-js';
import { SERVER_KEY, port } from './env';

const EVENTS = ['onmatchdata', 'onmatchpresence', 'onmatchmakermatched', 'onpartypresence', 'onpartydata', 'onparty', 'onpartyclose', 'onchannelmessage', 'onchannelpresence', 'onstatuspresence', 'onnotification', 'ondisconnect', 'onerror', 'onstreamdata', 'onstreampresence'] as const;

export type Ev = { type: string; ev: any; t: number };

export class Player {
  client!: Client;
  socket!: Socket;
  session!: Session;
  events: Ev[] = [];
  waiters: { type: string; pred: (e: any) => boolean; resolve: (e: Ev | null) => void; after: number }[] = [];
  disconnected = false;
  constructor(public tag: string, public node: number, public device: string) {}
  get id() { return this.session.user_id!; }

  async signIn() {
    this.client = new Client(SERVER_KEY, '127.0.0.1', String(port(this.node)), false, 10000);
    this.session = await this.client.authenticateDevice(this.device, true, this.tag.slice(0, 40));
  }

  async connect(node = this.node, status = true) {
    this.node = node;
    this.client = new Client(SERVER_KEY, '127.0.0.1', String(port(node)), false, 10000);
    const socket = this.client.createSocket(false, false);
    for (const e of EVENTS) {
      (socket as any)[e] = (ev: any) => {
        const item = { type: e, ev, t: Date.now() };
        if (e === 'ondisconnect') this.disconnected = true;
        this.events.push(item);
        for (const w of [...this.waiters]) {
          if (w.type === e && item.t >= w.after && w.pred(ev)) {
            this.waiters.splice(this.waiters.indexOf(w), 1);
            w.resolve(item);
          }
        }
      };
    }
    this.disconnected = false;
    await socket.connect(this.session, status);
    this.socket = socket;
    const me = await this.rpc('cle2e_whoami');
    if (me.node !== `node${node}`) throw new Error(`connected to ${me.node}, wanted node${node}`);
    return this;
  }

  async rpc(id: string, body: any = {}) {
    const res = await this.client.rpc(this.session, id, body);
    return res.payload as any;
  }

  count(type: string, pred: (e: any) => boolean = () => true) {
    return this.events.filter((e) => e.type === type && pred(e.ev)).length;
  }

  wait(type: string, pred: (e: any) => boolean = () => true, timeout = 8000, since = 0): Promise<Ev | null> {
    const hit = this.events.find((e) => e.type === type && e.t >= since && pred(e.ev));
    if (hit) return Promise.resolve(hit);
    return new Promise((resolve) => {
      const w = { type, pred, resolve, after: since };
      this.waiters.push(w);
      setTimeout(() => {
        const i = this.waiters.indexOf(w);
        if (i >= 0) {
          this.waiters.splice(i, 1);
          resolve(null);
        }
      }, timeout);
    });
  }

  close() {
    try { this.socket?.disconnect(false); } catch {}
  }
}

let seq = 0;
export const runTag = Date.now().toString(36).slice(-5);

export async function player(node: number, name = 'p') {
  seq++;
  const p = new Player(`cle2e${runTag}${name}${seq}`, node, `cle2e-${runTag}-${name}-${seq}-device-xx`);
  await p.signIn();
  await p.connect(node);
  return p;
}

export const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

export function joinsOf(e: any, userId: string) { return (e.joins || []).some((j: any) => j.user_id === userId); }
export function leavesOf(e: any, userId: string) { return (e.leaves || []).some((j: any) => j.user_id === userId); }
