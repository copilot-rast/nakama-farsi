# Cluster spike: doublemo/nakama-plus (2026-11-11)

Goal: before phase 6 ports the nakama-plus cluster layer onto this fork, check on real processes whether two players on different nodes can play, match, chat and see each other online, what happens when a node dies, and how the design scales to the owner's target (about 30 pods, about 1M concurrent players).

Short answer: **every feature works across nodes in the simple case, but the replication layer is lossy and does not scale.** Under load the nodes disagree about who is online and who is in which match, leaves are not replicated, every node matches the same tickets, and the gossip channel is already full at 5k players on 4 nodes. Port the routing ideas, not the replication layer.

## Setup

- Code: `doublemo/nakama-plus` at `a75f67c` ("Prepare 3.41.0 release"), shallow clone, built with Go 1.27.2 from its vendor folder. 4 nodes as plain processes on this Mac (14 cores, 36 GB), no Docker.
- Ports per node `i`: grpc `76i0`, socket `76i1`, console `76i2`, gossip `76i3`, Prometheus `76i4`.
- Database: separate `nakama_plus_spike` on the shared Postgres 5481, dropped at the end.
- Clustering: hashicorp/memberlist v0.7.0 (UDP gossip + TCP). Discovery is either etcd (`cluster.etcd.endpoints`) or a static `cluster.members` list. No etcd was installed, so static members were used. Redis is not used.
- Runtime: one Lua module (relay authoritative match, RPCs `create_auth_match`, `notify_user`).
- Clients: nakama-js 2.x in Node 22, each player's socket pinned to one node's port. Load: k6 1.x, `k6/ws`.

Config facts found while starting:

1. With a full static member list, the **first node cannot start** (`failed to join cluster`, fatal, because no peer is up yet). One node must start with an empty list (a seed) and the rest list the seed.
2. With the default config (`retransmit_mult` not set) presence updates reach peers only by luck: the delegate's `TransmitLimitedQueue` uses `RetransmitMult: 0`, so every update is sent once, to one random node, and the rest wait for the 30 s push/pull. Measured: a new relayed match on node2 became joinable from other nodes after 0.1–1 s most of the time, but 9.9 s and 27.1 s in 2 of 15 tries. With `cluster.retransmit_mult: 4` it was 0.1 s every time. **All results below use `retransmit_mult: 4`.**
3. `push_pull_interval` is copied into `ProbeInterval` (bug in `toMemberlistConfig`), so push/pull cannot be tuned.

## The four questions (player A on node2, player B on node3)

Each check was also run with all players on node2; that baseline passed 47/47, so every failure below is cross-node only.

### 1. Can they see each other in a match?

| Check | Result | Evidence |
|---|---|---|
| Relayed match: B joins A's match right after A creates it | FAIL | `Match not found`; works after about 100 ms (3 tries), once the presence gossip arrives. Under 5k load it took about 0.5 s. |
| Relayed match: A sees B's join | PASS | |
| Relayed match data A→B / B→A | PASS | p50 0.55 / 1.0 ms, p95 3.1 / 1.6 ms (same node: p50 0.33–0.36 ms) |
| Authoritative match (host node2 or node3, guest on the other) join + data both ways | PASS | p50 about 78 ms both ways (bound by the 10 Hz tick; same as same-node) |
| Leave (relayed, authoritative, party, room channel): other side sees it | **FAIL** | No leave event in 5 s |
| After leaving, the leaver stops getting the match/party data | **FAIL** | Leaver still got 3/3 messages of a match/party it had left |

Cause, in code: `LocalTracker.Untrack` builds the lookup key with the local node name (`PresenceID{Node: t.name, ...}`) even when it is called for a remote presence, so a replicated leave never finds the presence. Every remote node keeps a ghost presence until the player disconnects. One-line bug, but it shows the remote paths are not tested.

### 2. Can the matchmaker match them? Parties?

| Check | Result | Evidence |
|---|---|---|
| Tickets added on node2 and node3 matched together, both join by token and see each other | PASS | 0.2–1.3 s (interval 2 s) |
| Party created on node2, joined from node3, party data both ways | PASS | p50 0.6 ms |
| Party matchmaking: party (node2 leader + node3 member) + solo on node4 | PASS | 1.3 s |
| Party leave seen on the other node / leaver stops getting party data | FAIL | same Untrack bug |
| **Each player matched once** (8 players on 4 nodes, pairs, 5 rounds) | **FAIL** | 40 players got 80 matched events, 20 players matched twice, 40 distinct matches instead of 20. Even with all 8 players on node2: 58 events, 15 double-matched. |

Cause: a replicated ticket is inserted on every node with `matchmaker.Add`, which makes a new ticket id, marks it as owned by the local node and active. So every node runs the matchmaker over its own copy of every ticket and the same player is put into several matches. Removing one ticket by id from another node also misses, because the ids differ. In the 5k load run there were 31k matched events for about 21k tickets.

### 3. Can they chat?

| Check | Result | Evidence |
|---|---|---|
| Room channel A↔B, join event | PASS | p50 1.7–1.9 ms (same node 0.9 ms) |
| Group channel A↔B | PASS | p50 1.4–2.0 ms |
| Direct message A↔B (same channel id on both sides) | PASS | p50 1.9–2.0 ms |
| History from each node | PASS | both sides see all 20–21 messages (history is in the database) |
| Room leave seen by the other side | FAIL | Untrack bug |

### 4. Can they see each other online?

| Check | Result | Evidence |
|---|---|---|
| `listFriends` on node2 shows B (node3) online | PASS | |
| `status_follow` on node2 returns B's presence | PASS | |
| B changes status (`updateStatus`) → A gets the event | **FAIL** | never arrives. `UpdateTrackPeer` broadcasts `BinaryLog_UpdateTrack`, but `handleBinaryLog` has no case for it, so status text and any presence meta update never replicate. |
| B disconnects → A sees offline / B reconnects on node4 → A sees online | PASS | 60–160 ms |
| REST `getUsers` `online` flag for users on other nodes | FAIL under load | 35–39 of 100 connected users shown online (it looks at local data) |

### Also

| Check | Result | Evidence |
|---|---|---|
| Friend-request notification A(node2) → B(node3), live | PASS | |
| `nk.notification_send` from an RPC on node4 → B on node3 | PASS | 2.3 ms including the RPC |
| Logout on node2 invalidates the token on node3 | **FAIL** | token still valid on node3. The session cache (logout, ban) is node-local. |

## Killing a node

Players A (node2) and B (node3) in a relayed match on node2, an authoritative match on node3, a room channel, status follow, plus matchmaker tickets. node3 killed with `kill -9 <pid>`.

| What | Result |
|---|---|
| Failure detection | about 6 s ("Node left cluster" on all peers, memberlist suspicion) |
| B's socket | closed at once; B reconnected to node4 with the same token in 3 ms, A saw B online again |
| A's relayed match / room / status | **no leave events at all** (waited 30 s). `ClearTrackByNode` drops the dead node's presences silently, so opponents are never told the player left. |
| Authoritative match hosted on node3 | gone; A gets no "match ended" message, only `Match not found` on rejoin |
| `listFriends` on node2 | B offline (correct) |
| Matchmaker tickets of node3 | removed from the pool |
| Rest of the cluster | kept working |
| Restart node3 (seed list = other nodes) | rejoined in under 0.1 s, its new matches visible on all nodes in 0.1–0.2 s |

## Load

k6 sessions spread evenly over 4 nodes. Each session: device auth, socket with `status=true`, status update, follow 10 users, join one of 100 rooms, a matchmaker ticket (4 players, 10 modes) every 30–40 s, a room message every 60–90 s. Logger at WARN. Latency probe: relayed match, node2→node3 vs node2→node2, 10 msgs/s.

| | 1k (250/node) | 5k (1250/node) | 10k (2500/node) |
|---|---|---|---|
| RSS per node (max) | 103–106 MB | 258–267 MB | 406–439 MB |
| Go heap in use (max) | 42–48 MB | 156–166 MB | 253–267 MB |
| CPU per node (avg in hold) | 4–6 % | 16–18 % | 34–52 %, peaks 120 % |
| Goroutines per node | about 855 | about 3,800 | about 5,900 |
| Gossip UDP per node (in+out) | 15–48 KB/s (with ramp) | 28–30 KB/s | 26–27 KB/s (flat: saturated) |
| Peer TCP connections in TIME_WAIT (all nodes) | — | about 7,300 (≈ 245 new TCP connections/s) | port exhaustion |
| Presences per node (should be equal) | — | 10,767 / 15,009 / 15,009 / 10,767 | — |
| Connected users shown online, by node | — | 232–300 of 300 (up to 23 % missing) | — |
| Ghost presences 45 s after all disconnected | 1,270–1,750 | 7,490–11,241 | 5,231–5,403 |
| Disconnected users still shown online | 34–44 % | 50–75 % | 0–25 % |
| Cross-node probe latency | p50 0.42 ms, p95 0.88 ms, 0 lost | p50 0.41 ms, p95 2.0 ms, **8 of 401 lost** | probe could not connect |
| Same-node probe latency | p50 0.33 ms, p95 0.76 ms | p50 0.34 ms, p95 1.9 ms, 0 lost | — |
| k6 errors | 0 | 0 | 13,912 |
| Server warnings | — | 673 × "Binary log broadcast channel full, dropping message" | 2,855 × that, 22,850 × "Failed to send broadcast to node: can't assign requested address" |

The 10k run failed on this one Mac: k6's 10k sockets plus the peer layer's short TCP connections used up the loopback ephemeral ports (16k). Nodes then missed acks, suspected each other (58 suspect events) and dropped each other's whole state. Part of this is the single machine, but the TCP churn comes from the design (below).

## How the design works (from the code)

- **Full replication.** Every presence (status, channels, matches, parties, notifications stream), every matchmaker ticket, party index entries and leaderboard ranks are copied to every node. Each change goes out as a "binary log" message.
- **Transport for replication:** memberlist's gossip queue (UDP packets of 1,400 bytes, 3 random nodes every 200 ms). Receivers do not forward. With `retransmit_mult: 4` each message is sent `4 × ceil(log10(N+1))` times: 4 at 4 nodes, 8 at 30 nodes.
- **Ordering:** one version (a timestamp) per sending node, and a receiver drops anything older than the last version it saw from that node. The gossip queue sends the freshest messages first, so under load older messages (often the leave) arrive after newer ones and are thrown away. Plus a 16-slot channel before the queue drops messages when busy (673–2,855 drops seen). This is why nodes diverge and ghosts stay.
- **Repair:** push/pull every 30 s with one random node sends the node's whole tracker (all cluster presences) over TCP, but only replaces state when the version is newer, and memberlist rejects push/pull state over 20 MB.
- **Routing (the good part):** messages for a presence on another node (match data, chat, notifications, party data) go only to the nodes that host the recipients. But each message is a memberlist `SendReliable`, which **opens a new TCP connection per message**; match/party calls to a remote host use `Request`, which is two of them plus a wait.
- **Leader:** only exists with etcd (election), and nothing uses it except the status page. Leaderboard/tournament schedulers run on every node, so reset/end hooks fire once per node.
- Not cluster-aware or broken across nodes: session invalidation (logout/ban, tested), status updates (tested), leaves (tested), matchmaker ownership (tested), dead-node leave events (tested), REST `online` flag (tested), scheduler hooks (code), the Lua/JS `cache*` runtime functions (need etcd, not tested). Storage, wallets, groups, friends and chat history are fine because they live in the database. Notifications work (tested).
- It depends on forks: `doublemo/nakama-common` and `doublemo/nakama-kit` (22 server files import the kit's `pb`). Our fork keeps `heroiclabs/nakama-common`, so the port means rewriting those types, not copying.

## Scaling estimate for 30 pods and 1M players

Assumptions: 1M players, about 33k per pod, 3 presences per player (what the load run showed: 15k presences for 5k players), presence changes about once per 30 s per player, about 200 bytes per replicated record.

- **Memory per pod.** Measured: 1250 local sessions + about 15k replicated presences/tickets = about 260 MB; going from 1k to 5k total added about 150 MB per node. With full replication every pod holds all 3M presences plus every ticket in its own bluge index. At 1–5 KB per replicated item (Go maps, index docs) that is **3–15 GB per pod just for copies of other pods' state**, on top of its own 33k sessions (about 3 GB at the measured 80–100 KB per session). Adding pods does not help: each new pod also holds everything.
- **Gossip.** Cluster-wide about 100k presence changes/s (plus tickets). Each pod originates about 3.3k/s, each must be sent 8 times: about 5 MB/s outbound per pod. memberlist's gossip queue moves at most about 21 KB/s out per node (3 × 1,400 B × 5/s), and we already saw it flat at about 29 KB/s in+out with only 5k players. So the queue grows without limit, and even when it drains, 8 random sends reach only about 7 of 29 peers (about 25 %). Most pods would show stale presence all the time.
- **Push/pull repair.** 3M presences × 200 B = about 600 MB of state, 30× over memberlist's 20 MB cap: repair and the initial sync for a new or restarted pod fail completely. The cap is hit at about 100k presences, i.e. about 30k players in the whole cluster.
- **Cross-node routing.** With 30 pods and random placement, about 97 % of match/party/chat partners are on another pod. One TCP connection per message: on Linux a pod pair can open about 28k connections per 60 s TIME_WAIT window, i.e. **about 470 messages/s per pod pair**. One authoritative match at 10 Hz with 4 remote players already uses 40/s; a pod with 33k players would need tens of thousands per second per peer.

Where it breaks first: the gossip/version replication (wrong online and match state already at 5k players on 4 nodes), then the 20 MB push/pull cap (about 30k players cluster-wide), then per-message TCP routing, then memory.

## Risks if ported as is

- Wrong presence everywhere: ghosts online, missing players, players who left still getting match data (security issue for authoritative matches: data from a match they left).
- Players matched into several matches at once.
- Logout and bans do not reach other pods.
- No leave events when a pod dies; matches on a dead pod vanish with no message.
- Tournament and leaderboard reset hooks run 30 times.
- Large dependency change (two forked libraries) and a memberlist/etcd ops surface.

## Recommendation for phase 6

1. Do not port the replication layer (binary log over gossip, full copies of tracker/matchmaker/rank cache). It cannot reach the target and is wrong at small scale.
2. Keep the ideas that work: node id in every presence id and match id (`uuid.node`), routing a message only to the nodes that host recipients, remote match join/data/signal by asking the host node, party calls to the leader's node, and memberlist only for membership and failure detection (with `retransmit_mult` set and a seed list), or Kubernetes endpoints for discovery.
3. Replace the rest with sharded ownership instead of full copies:
   - Presence/status: each user's presence lives on one owner (hash of user id over the pods, or a Redis/NATS-style store with TTL heartbeats). Followers subscribe to the owner, not to everyone.
   - Matchmaker: one owner per pool shard (e.g. by mode/region hash); tickets are sent to the owner, which is the only one to match them.
   - Cross-node messages over long-lived gRPC streams (one per pod pair) with batching, not a TCP connection per message.
   - Session invalidation (logout/ban) as a cluster broadcast; leaderboard/tournament schedulers under our `clusterkit` lease (one node runs them).
   - On node death: emit leave events for its presences and a "match ended" for its matches.
4. Plan the port as our own small cluster package on `heroiclabs/nakama-common`, using nakama-plus as a reference for the touch points (tracker, message router, match registry, party registry, matchmaker, session registry), and keep the multi-node test scripts from this spike as the e2e base.
5. Before building, a short decision is needed from the owner on the shared store for presence (Redis vs owner-pod hashing), since that sets the ops cost.

## Not tested

- etcd discovery and leader election (etcd not installed); the Lua/JS `cache*` functions that need it.
- More than 4 nodes, and nodes on separate machines (all traffic was loopback on one Mac).
- Leaderboard rank replication and tournaments across nodes (read from code only).
- 10k sessions cleanly (the single Mac ran out of ports).
- Long soak, network partitions, clock skew between pods (versions are wall-clock timestamps).
