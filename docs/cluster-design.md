# Cluster design (phase 6)

Owner's decision (2026-11-11, after `docs/cluster-spike.md`): do not port nakama-plus's replication. Take its routing ideas only and build the shared state on Redis. Target: about 30 pods and about 1M concurrent players. This document is the contract for the code in `internal/cluster` and `server/cluster_*.go`.

## Goals and rules

1. **Off by default.** With `cluster.enabled: false` (the default) `main.go` builds exactly the same local components as upstream Nakama 3.42; every cluster hook in the local types is a nil field that is never called. A single node behaves as before, byte for byte on the wire.
2. **Same module path and libraries.** `github.com/heroiclabs/nakama/v3`, `heroiclabs/nakama-common` unchanged, so the five plugins and the console overlays load as before. One new dependency: `github.com/redis/go-redis/v9` (vendored). No new protobuf: the inter-node transport uses gRPC with a raw-bytes codec and its own small frame format.
3. **Main Postgres untouched.** The only database use is the existing `cluster_lease` table from phase 0b (`internal/clusterkit`) for scheduler claims.
4. **Each node holds only what it needs**: its own sessions, its own presences, and for every stream that has at least one local presence a view of that stream's other members. No node holds a copy of every presence, ticket or match.
5. **Redis work per event is O(1)** (one script call, pipelined with the node index write) and inter-node messages per event are O(nodes that host a recipient). Messages to a stream never touch Redis when the sender hosts the stream.
6. **No TCP connection per message.** One gRPC connection per node pair (HTTP/2, multiplexed): one long-lived client stream for one-way frames (batched) plus unary calls on the same connection for request/response.

## Components

| Piece | Where | What |
|---|---|---|
| `cluster.Redis` | `internal/cluster/redis.go` | go-redis universal client (single, Sentinel or Redis Cluster addresses), key prefix, Lua scripts |
| Membership | `internal/cluster/membership.go` | node registry, heartbeats, death detection, reap leases, epoch |
| Transport | `internal/cluster/transport.go` | gRPC server + one peer client per node, frame batching, unary calls, shared-secret auth |
| Presence store | `internal/cluster/presence.go` | stream hashes, node counts, versions, follower sets, node indexes |
| Matchmaker shards | `internal/cluster/shards.go` | shard leases, rendezvous preference, fenced ticket store |
| Invalidation log | `internal/cluster/invalidation.go` | Redis stream of logout/ban records, tail + replay |
| Signals | `internal/cluster/signal.go` | small pub/sub used by plugins (live support chat) |
| Server glue | `server/cluster*.go` | wrappers that implement `Tracker`, `MessageRouter`, `SessionRegistry`, `SessionCache`, `StatusRegistry`, `StreamManager`, `MatchRegistry`, `PartyRegistry`, `Matchmaker`, `StatusHandler`, and the frame/call handlers |

`main.go` creates the local components as today and passes each through `cluster.X(local)`. With clustering off `server.NewCluster` returns nil and every `(*Cluster)(nil).X(local)` returns `local` unchanged.

## Node identity and membership

- Node name: `--name`, unique per process (Kubernetes: the pod name). Session IDs already carry the first 6 bytes of `sha1(name)` (upstream `NodeToHash`), so the node of any session is known from its ID; match and party IDs are `uuid.node`. These are the routing keys.
- Incarnation: a random 8-byte id made at start. A restarted pod with the same name is a new incarnation; the old one is reaped.
- Registry (Redis):

| Key | Type | Value | TTL |
|---|---|---|---|
| `nk:nodes` | hash | field `<name>/<inc>` → JSON `{addr, hash, started}` | none, removed after reap |
| `nk:hb:<name>` | string | `<inc>` | `node_ttl_ms` (default 5000), refreshed every `heartbeat_ms` (1000) |
| `nk:epoch` | string | random id set once with `SET NX` | none |
| `nk:reap:<name>/<inc>` | string | reaper holder | 30 s, renewed while reaping |

- A node is **alive** when `GET nk:hb:<name>` equals its incarnation. Every node runs a watcher every `heartbeat_ms`: `HGETALL nk:nodes` + `MGET` of the heartbeat keys (30 pods = 2 commands). Alive entries it did not know → peer up (dial). Entries no longer alive → peer down.
- **Death detection** = heartbeat key expired or replaced by a new incarnation: about `node_ttl_ms` (5 s) after a crash, at once after a restart with the same name, at once after a graceful stop (the node deletes its own keys).
- **Redis outage safety**: a node judges others only after its own heartbeat writes have succeeded continuously for one full `node_ttl_ms`. After a Redis outage every node refreshes first, so no live node is declared dead.
- **Epoch**: on every heartbeat a node also reads `nk:epoch`. If it is missing or changed (Redis restarted empty), the node writes a new one if missing and republishes all its local presences, follows and owned matchmaker tickets. Redis therefore needs no persistence.
- **Self reaped** (a long pause made others reap this node): the heartbeat sees a foreign incarnation in `nk:hb:<name>`; the node takes a new incarnation, re-registers and republishes its state.
- Why Redis and not memberlist or Kubernetes endpoints: Redis is already required for presence and matchmaking, a heartbeat key is one command per second per node, failure detection is deterministic (no gossip suspicion), the same code runs in Docker Compose, on bare processes and in Kubernetes, and there is no UDP/gossip port to open. A headless Kubernetes service is not needed for membership.

## Transport

- Each node listens on `cluster.port` (default 7353) with its own gRPC server (service `nakama.cluster.Peer`, raw-bytes codec). Peers authenticate with `cluster.secret` (default: `session.encryption_key`) in metadata, compared in constant time; the advertised address is `cluster.advertise_address:cluster.port`.
- Per peer: one `grpc.ClientConn` (one TCP connection). One-way frames go over one client stream `Pipe`; a writer goroutine drains the peer queue and packs up to 256 frames / 256 KB into one gRPC message (batching under load, no delay when idle). Request/response goes over unary `Call` on the same connection.
- Frame: `kind (1 byte) | len (uvarint) | payload`. Hot frames (envelopes, presence ops, match data) use a compact binary encoding; calls use JSON.
- Queue full (default 65,536 frames) → frame dropped, counted, and the peer is asked to resync its views after the queue drains. Stream broken → reconnect with backoff; both sides resync their views that contain the other node (frames in flight may be lost).
- Ordering: frames from A to B arrive in send order. Unary calls are independent.

## Presence and streams in Redis

A stream is `(mode, subject, subcontext, label)`. Its key `sk` is `mode|subject|subcontext|label`. Per-user streams (notifications mode 0, status mode 1) use the hash tag `{u:<user>}` so the status stream and the user's follower set share a slot; every other stream uses `{s:<sk>}`. So every script touches one slot and works on Redis Cluster too.

| Key | Type | Content |
|---|---|---|
| `nk:{tag}:p` | hash | field `<node>|<session>|<user>` → meta `inc|format|hidden|persistence|username|status` |
| `nk:{tag}:n` | hash | field `<node>` → number of that node's presences in the stream; field `_v` → stream version |
| `nk:{u:<user>}:f` | hash | field `<node>` → 1 when that node has at least one session following the user |
| `nk:np:<name>/<inc>` | set | `sk` + session + user of every presence of that incarnation (for reaping and reconcile) |
| `nk:nf:<name>/<inc>` | set | users followed by that incarnation |

No TTLs: entries are removed by the owning node on leave and by the reaper on death; every key disappears when empty.

**Join / update** (one Lua script, plus `SADD` on the node index in the same pipeline): `HSET p field meta`; if the field is new, `HINCRBY n node 1`; `HINCRBY n _v 1`; return `{new, version, nodes, followers?, snapshot?}`. `snapshot` (the whole `p` hash) is returned only when the caller has no view of the stream yet. **Leave**: `HDEL`, decrement, drop the node field at 0, bump the version, return the removed meta and the nodes. Reads: `HGETALL p` (list), `HLEN p` (count), `EXISTS p`, `HKEYS n`.

### Views and events

- The local tracker (upstream `LocalTracker`) still stores this node's presences and still delivers presence events to local sessions. A new `clusterHook` field (nil when off) gets every local join, update and leave, hidden ones included.
- For every stream with at least one local presence the node keeps a **view**: the other nodes' presences in that stream and the snapshot version. It is created from the snapshot returned by the first local join and deleted when the last local presence leaves.
- After the Redis script the node sends one presence frame per target node: `stream, version, ops` (join/update/leave with the full presence; a `deliver` flag that is false for hidden presences and for silent removals such as a closed match). Targets = nodes in `n` except self, plus the host node of an authoritative match or party (its label), plus, for a status stream, the follower nodes from `f`.
- A receiving node updates its view (op version must be newer than the view's snapshot version; ops that arrive while the view waits for its snapshot are buffered and replayed) and, for `deliver` ops, injects the event into its local tracker queue. The upstream `processEvent` then does the rest unchanged: channel/match/party/stream presence events to local sessions, match and party join/leave listeners on the host, status events to local followers.
- Ordering argument: the Redis script serialises all changes of one stream, so a joiner either sees a presence in its snapshot or is in `n` when that presence changes later and gets the op. Ops about one presence always come from the same node (or from the reaper after that node died), so they arrive in order.
- Tracker reads: with a view, `ListByStream`, `CountByStream`, `StreamExists`, `ListPresenceIDByStream`, `ListNodesForStream` are answered from memory (local + view). Without a view (for example an RPC sending to a user on another node) they read Redis. `GetLocal…` and `ListLocal…` stay local. `CountByStreamModeFilter` (relayed match listing) asks every node for its local counts and sums them.
- `UntrackByStream` (match end, stream close, group delete): `HKEYS n`, then every hosting node removes its local presences silently.

### Status, followers, online

- `Follow`: the local status registry calls a hook when a user gets its first local follower (`HSET f node 1`, `SADD nf`) or loses its last (`HDEL`, `SREM`).
- A status presence change is sent to the stream's nodes and to the follower nodes; each delivers it to its local followers through the unchanged `LocalStatusRegistry`.
- `IsOnline` / `FillOnline…` (REST `online` flags for users, friends, group members, accounts): one pipelined `EXISTS nk:{u:<user>}:p1` per user (status stream not empty), so the flag is cluster-wide and exact.

## Message routing

| Call | Cluster behaviour |
|---|---|
| `SendToPresenceIDs` | local IDs → local router; the rest grouped by `PresenceID.Node` → one envelope frame per node (`reliable`, session IDs, envelope bytes) |
| `SendToStream` | local sessions of the stream + one stream frame per other node in the view (or `HKEYS n` without a view); the receiver delivers to its local sessions in that stream |
| `SendToAll` | local + one frame to every alive node |
| `SendDeferred` | as `SendToPresenceIDs` per message |
| `NotificationsSend` (live part) | recipients' notification-stream presences from views or one pipelined `HKEYS` per user, then `SendToPresenceIDs` |
| Relayed match data | membership check is local (the sender's own presence); recipients come from the view (no Redis per message) |

The envelope is marshalled to protobuf once per target node; the receiver re-encodes it per session format (JSON or protobuf) exactly as the local router does.

## Sessions, logout, ban, suspension

- `SessionRegistry.Disconnect(session)`: local if the session is here, else one disconnect frame to the node found from the session ID hash (with the ban notice when `ban`).
- `SingleSession`: local plus one frame to every node that hosts the user's notification stream.
- `SessionCache` (`Remove`, `RemoveAll`, `Ban`, `Unban`): applied locally, then appended to the Redis stream `nk:inv` (`XADD … MINID ~ now-maxTokenLifetime`). Every node tails it with a blocking `XREAD` on its own connection and applies each record to its local cache; at start a node replays the records newer than the longest token lifetime. So a logout or ban on one node invalidates the token on every node within milliseconds, and a node started later still knows it. The console session cache uses its own stream `nk:invc`.
- Our plugins need no change: phase 3 suspension (`SessionLogout` + `StreamUserList(0, user)` + `SessionDisconnect`) and jazebe's suspension now reach every node through the three rows above. The phase 3 `restrictionLoop` stays as a database-side safety net.
- Live notifications (publication in-app, support 1210/1211, any `NotificationsSend`) are routed to the node that holds the player's socket.

## Matchmaker: shards with exactly-once matching

- Tickets are split into `cluster.matchmaker_shards` shards (default 16) by `fnv32(stringProperties[cluster.matchmaker_shard_property]) % shards`. With no property configured every ticket lands in the same shard, i.e. one node matches for the whole cluster (one node's matchmaker handles tens of thousands of tickets per interval). Games that want to spread matching set a pool property such as `mode` or `region`; tickets only match within their shard.
- Ownership: `nk:mm:{<i>}:own` = `<name>/<inc>/<token>` with `PX` 6 s, renewed every 2 s by a fenced script. Preferred owner of shard `i` = rendezvous hash of `i` over the alive nodes; a node acquires free shards it prefers (or any shard free for more than 3 watcher rounds) and hands a shard to its preferred owner once that node has been alive for 10 s (stop processing, release, the other side loads).
- Each owned shard is a full upstream `LocalMatchmaker` instance (own bluge index, own interval ticker, same runtime hooks: matchmaker override/processor functions, matched hook). The owner is therefore the host of any authoritative match created by the matched hook.
- Tickets are stored in `nk:mm:{<i>}:t` (hash ticket → JSON `MatchmakerExtract`). `Add` on any node is forwarded (unary call) to the shard owner, which runs `LocalMatchmaker.Add` and writes the ticket with a fenced `HSET` (fails if it lost the lease). The origin node remembers ticket → shard for its sessions and parties to route removals.
- **Commit point**: after `Process` forms matches and before any `MatchmakerMatched` is sent, a fenced script checks that the owner token is still ours and that every ticket of the match is still in the hash, then deletes them. Only committed matches are delivered. Lost lease → nothing delivered (the new owner still has the tickets). A ticket removed meanwhile → that match is not delivered, its other tickets are put back. So every ticket is matched at most once and never after its removal, even across failover.
- Removals (`RemoveSession`, `RemoveSessionAll` on disconnect, `RemoveParty*`) delete the ticket from Redis directly (any node may) and forward to the current owner. Unknown ticket → all shards.
- New owner (failover or handover): loads the shard hash with `Insert`, dropping tickets whose presences or party host are on dead nodes. Node death: every owner removes that node's tickets from its shards.
- Not available: no owner yet (failover window ≤ 6 s) → `Add` returns `ErrMatchmakerNotAvailable`, as upstream does when the matchmaker is stopped.
- Stats: `GetStats` sums the node's own shards.

## Matches

- Authoritative matches run on the node that created them (`uuid.node`). The registry wrapper sends every call for another node's match to that node: `GetMatch`, `JoinAttempt`, `Signal`, `GetState` as unary calls; `SendData` as an ordered one-way frame. The joiner then tracks its own presence locally; the presence op reaches the host (its label is a target) and the upstream join listener queues the join in the match handler.
- The match handler sends to its presences through the cluster router (presence IDs carry the node). `MatchKick` of a presence on another node → untrack frame to that node.
- Match end: `RemoveMatch` → cluster `UntrackByStream`.
- Relayed matches have no host: the stream lives in Redis, `matchJoin` checks `StreamExists` there, data is routed from views.
- `ListMatches`: authoritative matches are listed by every node (one call each, same filters, merged, limit applied after merge); relayed match sizes are summed over all nodes. Cost O(nodes) per listing.
- **Host death**: every node removes its local presences from matches and parties labelled with the dead node (silently) and sends each of those sessions a non-persistent notification code `-100`, subject `match_ended` (content `{"match_id":…,"reason":"node_down"}`) or, for a party, the upstream `PartyClose` message.

## Parties

Same host routing: the party lives on the leader's node at creation (`uuid.node`). `PartyJoinRequest`, `Promote`, `Accept`, `Remove`, `Close`, `JoinRequestList`, `MatchmakerAdd/Remove`, `DataSend`, `Update` for another node's party are unary calls to the host. The party handler adds members through the cluster stream manager, which tracks a remote session on its own node (unary call to the node from the session ID). `PartyList` asks every node and merges (no cursor across nodes).

## Node death

1. Watchers see the heartbeat missing → peer down on every node: close the peer connection, end local presences in matches/parties hosted there (above), remove the dead node's tickets from owned shards.
2. One node wins `SET nk:reap:<name>/<inc> NX` and reaps: scans `nk:np:<name>/<inc>` in batches of 500, runs the leave script for each presence and sends leave ops (`deliver` unless hidden) to the stream's nodes and hosts, exactly as if those sessions had disconnected. Opponents get presence leave events, match and party hosts get leave listeners, followers get status leave events. Then it removes the follower entries in `nk:nf:…`, deletes the index keys and the `nk:nodes` entry.
3. A reaper that dies midway loses its lease after 30 s and another node continues; every step is idempotent (a leave op is only sent for a field that was really removed).
4. Matchmaker shards of the dead node are free when their 6 s lease expires; preferred owners take them over and load the tickets.

## Schedulers and per-node workers

- Leaderboard and tournament scheduler: every node still computes the schedule and trims its own rank cache, but each callback (`tournament_end`, `tournament_reset`, `leaderboard_reset`, including the `size = 0` reset) is claimed first with `clusterkit.AcquireLease("lbs:<kind>:<id>:<ts>", node, 1h)`; only the node that wins runs it. Old claim rows are swept daily by the winner of `lbs-sweep`.
- Leaderboard cache: create and delete on one node are broadcast; the other nodes refresh their cache and scheduler.
- Plugin workers already follow `docs/cluster-readiness.md` (SKIP LOCKED claims and `clusterkit` leases).

## Plugins: live support chat signal (phase 5)

`internal/cluster` exposes `Signal(topic)` and `WaitSignal(ctx, topic, max)`. With clustering on, signals travel over Redis pub/sub (`nk:sig:<topic>`, one subscriber connection per node); the base plugin publishes `chat:<id>` after every message, typing change, take, transfer, close and player online change, and the console long-poll waits on it for up to 5 s before re-reading the database (fallback). With clustering off `WaitSignal` is a plain 600 ms sleep, exactly today's loop.

## Failure behaviour

| Failure | Behaviour |
|---|---|
| Node crash | Detected in ≤ `node_ttl_ms` (5 s); leaves, status offline, match/party leaves on hosts, `match_ended` to players of its matches; its shards fail over in ≤ 6 s; its players reconnect to any node with the same token |
| Graceful stop (SIGTERM) | Sessions close normally (leave events), shards released at once, registry entry removed: other nodes react immediately |
| Restart with same name | Old incarnation reaped at once; new one registers |
| Redis down | Local play continues (local matches, local chat). Cross-node presence writes fail and are logged; matchmaking `Add` fails; no node is declared dead. When Redis returns, each node reconciles its index with its local tracker (missing joins added, stale entries removed with leave events) |
| Redis restarted empty | Epoch changes; every node republishes its presences, follows and owned tickets |
| Peer stream broken | Reconnect with backoff, both sides resync views that include the other node |
| Peer queue full | Frames dropped (logged and counted), peer told to resync |
| Long pause / partition of one node | Others reap it after 5 s; when it comes back it takes a new incarnation and republishes; its players on others' matches already got leaves |

## Configuration

All flags are also YAML keys under `cluster:`.

| Flag | Default | Meaning |
|---|---|---|
| `--cluster.enabled` | `false` | Turn clustering on |
| `--cluster.redis_address` | `127.0.0.1:6379` | Comma-separated `host:port` list (one = single Redis, several = Redis Cluster, with `redis_master_name` = Sentinel) |
| `--cluster.redis_password` | empty | Redis password (`redis_username` for ACL users) |
| `--cluster.redis_db` | `0` | Database number (single/Sentinel) |
| `--cluster.redis_master_name` | empty | Sentinel master name |
| `--cluster.redis_tls` | `false` | TLS to Redis |
| `--cluster.prefix` | `nk` | Key prefix (several clusters on one Redis) |
| `--cluster.address` | empty (all) | Listen address of the inter-node gRPC port |
| `--cluster.port` | `7353` | Inter-node gRPC port |
| `--cluster.advertise_address` | first non-loopback IPv4 | Address other nodes dial (Kubernetes: the pod IP) |
| `--cluster.secret` | `session.encryption_key` | Shared secret between nodes |
| `--cluster.heartbeat_ms` | `1000` | Heartbeat and watcher period |
| `--cluster.node_ttl_ms` | `5000` | Heartbeat lifetime = death detection time |
| `--cluster.call_timeout_ms` | `5000` | Unary call timeout |
| `--cluster.send_queue_size` | `65536` | Frames queued per peer |
| `--cluster.matchmaker_shards` | `16` | Matchmaker shards |
| `--cluster.matchmaker_shard_property` | empty | String property that selects the shard |

Every node must use the same `session.encryption_key`, `session.refresh_encryption_key`, `socket.server_key`, runtime env and modules (`docs/cluster-readiness.md` rule 5).

## Deployment

- **Docker Compose**: `docker-compose.yml` has an optional `redis` service under the `cluster` profile and passes `NAKAMA_CLUSTER_*` from `.env` into the flags; with the defaults in `.env.example` clustering stays off and Redis does not start. Scale with `docker compose --profile cluster up --scale nakama=3` behind a load balancer for HTTP/WebSocket.
- **Kubernetes**: a `Deployment` (or `StatefulSet`) with `POD_NAME` and `POD_IP` from the downward API; args `--name=$(POD_NAME) --cluster.enabled=true --cluster.redis_address=redis:6379 --cluster.advertise_address=$(POD_IP) --cluster.port=7353`; container port 7353 open between pods (NetworkPolicy: only from pods with the same label); `terminationGracePeriodSeconds` ≥ `shutdown_grace_sec` + 5 so the node can release shards and say goodbye; a normal ClusterIP/LoadBalancer service for 7350 (HTTP + WebSocket, no sticky sessions needed) and 7351 (console). No headless service is needed for membership; one can be added for debugging (`dig` the pods). Redis: a managed Redis or a Sentinel set; no persistence required.

## Not cluster-aware (and why)

| Thing | Why it stays per node |
|---|---|
| Leaderboard rank cache | Rebuilt from the database on start and updated by local writes; rank numbers from different nodes can differ slightly between restarts. Use `leaderboard.blacklist_rank_cache` for boards that need exact ranks on every node |
| Storage search index (`storage.index`) | Upstream in-memory index; only filled from local writes after start. Not used by our plugins |
| Runtime `localcache*` functions | Named per node by design |
| Console login lockout cache | Per node; an attacker spread over pods gets N× the attempts. Console sits behind the admin network |
| Metrics, logs | Per node, scraped per pod |
| Match handover | A match never moves; when its node dies the players get `match_ended` |
| Party listing cursor | Results from all nodes are merged; no cursor across nodes |
| Max tickets per session | Enforced per shard, not cluster-wide |
