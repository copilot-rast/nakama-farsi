import http from "k6/http";
import ws from "k6/ws";
import { sleep } from "k6";
import { Counter, Trend } from "k6/metrics";
import encoding from "k6/encoding";

const TARGETS = (__ENV.TARGETS || "127.0.0.1:7611").split(",");
const VUS = Number(__ENV.VUS || 1000);
const RAMP = __ENV.RAMP || "60s";
const HOLD = Number(__ENV.HOLD || 90);
const PREFIX = __ENV.PREFIX || "ld";
const SKEY = __ENV.SKEY || "defaultkey";
const errors = new Counter("nk_errors");
const authErrors = new Counter("nk_auth_errors");
const matched = new Counter("nk_mm_matched");
const chatRecv = new Counter("nk_chat_recv");
const statusRecv = new Counter("nk_status_recv");
const connectTime = new Trend("nk_connect_ms");
const chatLatency = new Trend("nk_chat_ms");

export const options = {
  scenarios: {
    load: { executor: "ramping-vus", startVUs: 0, stages: [{ duration: RAMP, target: VUS }, { duration: `${HOLD}s`, target: VUS }], gracefulRampDown: "5s", gracefulStop: "15s" },
  },
  setupTimeout: "120s",
  discardResponseBodies: false,
};

const auth = "Basic " + encoding.b64encode(SKEY + ":");
export default function () {
  const id = __VU;
  const target = TARGETS[id % TARGETS.length];
  const uname = `${PREFIX}${id}`;
  const res = http.post(`http://${target}/v2/account/authenticate/device?create=true&username=${uname}`, JSON.stringify({ id: `${PREFIX}-device-${id}-xxxxxxxx` }), { headers: { Authorization: auth, "Content-Type": "application/json", Connection: "close" }, timeout: "30s" });
  if (res.status !== 200) { authErrors.add(1); sleep(5); return; }
  const token = res.json("token");
  const t0 = Date.now();
  let cid = 1;
  const send = (s, o) => { o.cid = String(cid++); s.send(JSON.stringify(o)); };
  ws.connect(`ws://${target}/ws?lang=en&status=true&token=${token}`, {}, function (socket) {
    socket.on("open", () => {
      connectTime.add(Date.now() - t0);
      send(socket, { status_update: { status: "playing" } });
      const follow = [];
      for (let k = 1; k <= 10; k++) follow.push(`${PREFIX}${((id + k * 37) % VUS) + 1}`);
      send(socket, { status_follow: { usernames: follow } });
      send(socket, { channel_join: { target: `lobby-${id % 100}`, type: 1, persistence: false, hidden: false } });
      const mm = () => send(socket, { matchmaker_add: { min_count: 4, max_count: 4, query: `+properties.mode:m${id % 10}`, string_properties: { mode: `m${id % 10}` } } });
      mm();
      socket.setInterval(mm, 30000 + (id % 10) * 1000);
      socket.setInterval(() => send(socket, { channel_message_send: { channel_id: `2...lobby-${id % 100}`, content: JSON.stringify({ t: Date.now() }) } }), 60000 + (id % 30) * 1000);
    });
    socket.on("message", (m) => {
      if (m.indexOf('"error"') >= 0) errors.add(1);
      else if (m.indexOf("matchmaker_matched") >= 0) matched.add(1);
      else if (m.indexOf('"channel_message"') >= 0) {
        chatRecv.add(1);
        const i = m.indexOf('{\\"t\\":');
        if (i >= 0) { const t = parseInt(m.slice(i + 7)); if (t > 0) chatLatency.add(Date.now() - t); }
      } else if (m.indexOf("status_presence_event") >= 0) statusRecv.add(1);
    });
    socket.on("error", () => errors.add(1));
    socket.setTimeout(() => socket.close(), (HOLD + 300) * 1000);
  });
}
