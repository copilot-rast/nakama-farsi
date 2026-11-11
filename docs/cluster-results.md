# Cluster results (phase 6)

Measurements and test results for the cluster built from `docs/cluster-design.md`. Machine: this Mac (14 cores, 36 GB), every process on loopback, Postgres 16 on 127.0.0.1:5481 (database `nakama_cluster`, created for these runs), Redis 8 on 127.0.0.1:6495 (`maxmemory 512mb`). All five plugins loaded (base, economy, publication, jazebe, stargame), logger at WARN.

## Step 0: one node, no cluster (baseline)

Binary of `dev` at phase 5 (`de7a6c54`), started as a plain node on ports 7611/7612. k6 1.x (`e2e/cluster/load/load.js`): each virtual user signs in with a device id, opens a socket with `status=true`, sets a status, follows 10 users, joins one of 100 rooms, adds a matchmaker ticket (4 players, 10 modes) every 30–40 s and posts a room message every 60–90 s. Same workload as the nakama-plus spike, so the numbers compare. Connections are spread over three local addresses (127.0.0.1, the LAN and the Tailscale address); sign-in requests use `Connection: close` so the client side does not keep a second port per user.

| | 1k | 5k | 10k | 15k |
|---|---|---|---|---|
| RSS (max) | 235 MB | 749 MB | 1,269 MB | 1,946 MB |
| Go heap in use (max) | 117 MB | 405 MB | 778 MB | 1,280 MB |
| CPU (average while holding / max) | 21 % / 43 % | 75 % / 108 % | 174 % / 252 % | 257 % / 367 % |
| Goroutines | 3,094 | 14,987 | 26,137 | 30,119 |
| Presences | 3,000 | 15,000 | 30,000 | 45,000 |
| Socket connect p95 | 2 ms | 1 ms | 3 ms | 2 ms |
| Room message delivered p50 / p95 / max | 2 / 3 / 15 ms | 1 / 3 / 15 ms | 1 / 4 / 385 ms | 2 / **6,855** / 16,145 ms |
| Room messages delivered | 8,850 | 323,600 | 1,507,008 | 3,415,194 |
| Matchmaker matched events | 3,096 | 21,968 | 48,000 | 71,137 |
| Errors | 0 | 0 | 234 client sign-ins: `can't assign requested address` (first run without `Connection: close`) | 17 socket errors; 2,781 sockets reconnected after the server closed them (`session outgoing queue full`) |
| Server errors | 0 | 0 | 72 × `Presence event dispatch queue is full` | same, plus `Failed to deliver presence event` |

Reading: one node of this fork serves about **10k players of this workload** with headroom; at 15k it is still up but room messages queue for seconds and slow sockets are dropped. The limit is not CPU (2.6 of 14 cores) but the single presence-event goroutine (queue of 1,024 events) and the per-session outgoing queue, both driven by the room fan-out (100 rooms of 150 players, every join and message goes to 150 sockets). Memory is about 105–130 KB per player with all plugins loaded (nakama-plus with one Lua module: about 80–100 KB). The macOS loopback has 16,384 ephemeral ports in total, so about 15k sockets is also the most one Mac can open in a test.

## Go tests (Redis on 6495, `CLUSTER_TEST_REDIS=127.0.0.1:6495`)

| Package | Tests | What they prove |
|---|---|---|
| `internal/cluster` | `TestFrameRoundTrip`, `TestTransportOrderedFramesAndCalls`, `TestTransportRejectsWrongSecret`, `TestTransportReconnectSignalsReset`, `TestTransportQueueFullDropsAndAsksResync` | 5,000 frames arrive in order and in batches, calls and errors round-trip, a wrong secret is refused, a reconnect asks for a resync, a full queue drops and asks for a resync |
| | `TestMembershipUpDownDeadAndReap`, `TestMembershipRestartSameNameReplacesIncarnation`, `TestMembershipDoesNotJudgeRightAfterStart` | three nodes see each other; a graceful stop needs no reap; a crash is detected within TTL + 4 heartbeats and exactly one of five reapers wins; a reaped name registers under a new incarnation; a restart with the same name marks the old incarnation dead; no node is judged before a full TTL of good heartbeats |
| | `TestShardOwnershipAndFencedCommit`, `TestShardExactlyOnceWithFailovers`, `TestPreferredSpreadsShards` | only the lease holder writes and commits; a commit with a removed ticket fails and keeps the rest; 4 simulated owners grabbing one shard with expiring leases and random removals for 1.5 s: about 270 tickets committed, every one exactly once, never a removed one; shard preference is spread over 3, 4 and 30 nodes |
| | `TestSignalsOffSleepAndOnWake`, `TestInvalidationLogReplaysAndTails` | without a cluster a wait is a plain sleep; a signal from another node wakes the waiter at once; logouts are replayed from the log and new ones tailed |
| `server` | `TestClusterRoutesToRemoteSessionsStreamsAndAll` | 3 nodes: direct, stream and broadcast envelopes reach sessions on other nodes; console status lists every node |
| | `TestClusterPresenceJoinLeaveAndListing`, `TestClusterStatusFollowOnlineAndUpdate` | channel joins, leaves, rejoin and disconnect seen across nodes; listing, counts and IDs from views; empty streams leave no key; follows, online flags, status text and offline across nodes |
| | `TestClusterNodeDeathSendsLeavesAndCleansRedis` | after a crash the other node gets the room leave, the offline event and `match_ended` for a match hosted on the dead node in about 0.45 s (TTL 0.4 s in the test), and the registry entry is reaped |
| | `TestClusterViewsStayConsistentUnderChurn` | 60 sessions on 3 nodes joining and leaving 4 streams concurrently: every node's view equals Redis, and Redis is empty after everyone leaves |
| | `TestClusterMatchmakerMatchesEachTicketOnceAcrossNodesAndFailover` | 24 players on 3 nodes, 4 shards by `mode`: three rounds, every player matched exactly once per round; the shard owner is killed with tickets waiting: the survivors take the shards and every player is matched exactly once more; a removed ticket is never matched |
| | `TestClusterLogoutBanDisconnectAndStreamJoinAcrossNodes` | logout and ban on one node invalidate on another; a node started later replays them; a remote stream join/leave through the session's node; a ban disconnect reaches the session on the other node |

Upstream's own server tests: `go test ./server/` (194 test functions, plus the cluster ones) passed against a disposable database `nakama_srvtest` (created, migrated, dropped) with `TEST_DB_URL`.

## Multi-node e2e (`e2e/cluster`, 4 nodes + Redis)

Two consecutive full runs, 9 of 9 passed each time (about 36 s per run).

| Owner's question / check | Result |
|---|---|
| 1. See each other in a match | Relayed: B on node3 joins A's match on node2 right after it is created, both see each other, data both ways, leave seen, the leaver gets nothing after leaving. Authoritative match hosted on node2: a player on node4 joins, data both ways through the host, `MatchSignal` from node1 answered by node2, `MatchList` on node4 finds it, a kick removes the presence on node4 and it stops getting data |
| 2. Matchmaker | 12 players on 4 nodes, three rounds: every player matched exactly once per round, 6 distinct matches per round, the pair joins the same match. Party with leader on node1 and member on node2, data both ways, party ticket plus a solo player on node3 matched together; party leave seen and the leaver stops getting party data |
| 3. Chat | Room with players on nodes 1–3: join lists, join/leave events on every node, messages (Persian text) everywhere, nothing after leaving; direct message between node1 and node3; history from node2 |
| 4. Online / status | Follow returns the status text from another node; REST `online` true; status change event; offline event and `online` false after disconnect; online again after reconnecting on another node; friend list `online` across nodes |
| Node killed with `kill -9` (by PID) | Room leave, relayed-match leave, offline status and `match_ended` all reached the player on node2 in 4.5–5.6 s (default `node_ttl_ms` 5 s); `online` false at once; the dead node's matchmaker ticket never matched a new player; the killed player reconnected on node4 with the same token and was online again; node3 restarted and its new player chatted in the same room |
| Logout | `session_logout` on node1: the old token is refused on node2 (waited > 1 s after sign-in: tokens are checked at one-second precision, as upstream) |
| Ban | ban from node1: the socket on node3 is closed and the token refused on node4 |
| Notifications, streams | notification sent from an RPC on node4 reached the socket on node2 in about 1 ms; runtime `stream_user_join` of a session on another node, `stream_send`, `stream_user_kick`, `session_disconnect` across nodes |
| Phase 3 suspension | suspension created through the console API of node1 closed the player's socket on node3 in about 0.5 s and the token was refused on node2 |
| Phase 5 live support chat | console long-poll on node1 (waiting up to 5 s on the chat signal) returned 7–9 ms after the player's message on node2 (a 600 ms database poll would need up to 600 ms; without the signal the 5 s wait would have taken about 3.5 s) |

## Load: 4 nodes + Redis (step 7)

Same k6 workload as step 0 and the spike, sessions spread evenly over nodes 1–4 (each node on three local addresses), matchmaker shards 8 by the `mode` property, logger at WARN. A latency probe runs 10 msgs/s through a relayed match node2→node3 and node2→node2 during the hold. In the middle of the hold an online check signs in on every node and reads the `online` flag of 300 load players; 20 s after the end the same check runs again, then the Redis presence and ticket hashes are counted.

| | 1k (250/node) | 5k (1,250/node) | 10k (2,500/node) | 15k (3,750/node) |
|---|---|---|---|---|
| RSS per node (max) | 139–164 MB | 280–288 MB | 376–417 MB | 514–559 MB |
| Go heap in use (max) | 61–78 MB | 158–185 MB | 234–266 MB | 328–358 MB |
| CPU per node, average in hold (max) | 1–4 % (13 %) | 4–18 % (30 %) | 10–51 % (72 %) | 27–79 % (132 %) |
| Goroutines per node | 650–710 | 2,650–2,710 | 5,150–5,230 | 7,680–7,730 |
| Presences per node | 747–759 (own sessions × 3) | 3,744–3,759 | 7,497–7,509 | 11,250–11,256 |
| Redis memory / ops per second in hold | 3.4 MB / 254 | 11.2 MB / 860 | 17–20 MB / 1,605–1,644 | 26.6–28.5 MB / 4,071–10,582 |
| Inter-node TCP connections | 24 (2 per node pair, fixed) | 24 | 24 | 24 |
| Online flag of 300 connected players, each node | — | — | 300/300 on all 4 nodes | 300/300 on all 4 nodes |
| Online flag 20 s after everyone left | — | — | 0/300 on all 4 nodes | first run 38 and 19 of 300 on two nodes (fixed, see below); final build 0/300 on all 4 nodes |
| Presences / tickets left in Redis after everyone left | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| Cross-node probe p50 / p95 / lost | 0.36 / 0.83 ms / 0 | 0.26 / 0.44 ms / 0 | 0.23–0.74 / 0.57–7.5 ms / 0 | 0.29 / 1.37 ms / 0 |
| Same-node probe p50 / p95 | 0.38 / 0.84 ms | 0.28 / 0.51 ms | 0.26–0.64 / 1.8–5.0 ms | 0.31 / 2.37 ms |
| Room message delivered p50 / p95 / max | 1 / 2 / 10 ms | 1 / 2 / 8 ms | 1–2 / 4–10 / 38–134 ms | 2 / 7 / 425 ms |
| Room messages delivered | 8,830 | 323,500 | 1,507,643–1,508,241 | 3,393,349 |
| Matchmaker matched events | 3,120 | 22,160 | 48,608–48,648 | 73,044 |
| k6 errors | 0 | 0 | 0 | 0 |
| Server errors | 0 | 0 | first run: 10 × matchmaker add timed out at the end of the run (fixed, see below); second run: 0 | 4–16 × `Presence event dispatch queue is full` when 15k sockets closed at once (upstream queue of 1,024; raise `tracker.event_queue_size` on big nodes), 1 add refused for a closed session |

The 10k and 15k rows show two runs each (before and after the fixes below); the final-build 15k run: RSS 491–541 MB, chat p95 6 ms, probe cross-node p50 0.36 ms, 0 presences and 0 tickets left.

### Compared with one node and with the nakama-plus spike

| | One node (step 0) | nakama-plus, 4 nodes (spike) | This cluster, 4 nodes |
|---|---|---|---|
| 10k players | 1,269 MB, 174 % CPU, 72 lost presence events, chat p95 4 ms | could not run (port exhaustion from per-message TCP), 13,912 k6 errors | 376–417 MB and 10–51 % CPU per node, chat p95 4–10 ms, 0 errors |
| 15k players | chat p95 6.9 s, 2,781 sockets dropped | — | chat p95 7 ms, 0 sockets dropped |
| Presences per node | — | 10,767 / 15,009 / 15,009 / 10,767 at 5k (copies of everyone, nodes disagree) | exactly 3 × own sessions |
| Connected players shown online | — | 232–300 of 300 at 5k | 300 of 300 on every node at 10k and 15k |
| Ghost presences after everyone left | — | 7,490–11,241 at 5k | 0 |
| Matched events / tickets | — | 31k events for about 21k tickets at 5k (double matching) | one event per player per match (48.6k at 10k = same as one node) |
| Inter-node traffic | memberlist gossip saturated at about 29 KB/s, 245 new TCP connections/s at 5k | — | 2 long-lived connections per node pair |

### Problems found by the load runs and fixed

1. **Matchmaker adds lost at the end of a run** (10k, first run): when a client closed while its add was in flight, the call was cancelled although the owner had already stored the ticket, so 10 tickets were left without a session. Fixes: calls to other nodes go over their own connection (frames can no longer delay a call behind HTTP/2 flow control), an add is not cancelled by the client going away and is undone if the session closed meanwhile, and owners drop tickets whose session's notification presence is gone (every 30 s). Second 10k run: 0 errors, 0 tickets left.
2. **Online flags after a disconnect storm** (15k): 38 and 19 of 300 players still showed online on two nodes 20 s after the run, while Redis had no presence. The flag fell back to the local status cache, which is fed by the presence event queue that drops events when 15k sockets close at once (`Presence event dispatch queue is full`, upstream behaviour). The flag now comes only from Redis.
3. **Unowned matchmaker shards for a moment at start-up** (first e2e run): a handover left a shard without owner for up to a second and adds failed; adds now retry until a new owner answers.

## Estimate for 30 pods and 1M players

Assumptions: about 33k players per pod, the same per-player activity as this workload (status, 10 follows, one room of about 150 players, a ticket every 30–40 s, a room message every 60–90 s), players spread evenly, all five plugins loaded. Measured on 4 nodes between 2,500 and 3,750 players per node.

| Resource | Measured slope | 30 pods × 33k | Notes |
|---|---|---|---|
| Memory per pod | about 120 KB per player + about 90 MB | **about 4.1 GB** | Only own sessions and views of own streams; adding pods does not add memory to the others (the spike's design needed 3–15 GB per pod just for copies) |
| CPU per pod | about 42 % of one core per 3,750 players (rooms of 150) | **about 3.7 cores** | Scales with message fan-out, not with cluster size; plan 6–8 vCPU pods or more, smaller pods |
| Redis memory | about 1.8 KB per player | **about 2 GB** | Presences, follower sets, node indexes, tickets |
| Redis operations | 0.16–0.7 per player per second | **about 160k–700k ops/s** | Mostly connects, status, matchmaker adds and commits; no Redis work per chat or match message. Above what one Redis primary does comfortably with Lua scripts: use Redis Cluster with 8 or more primaries (all keys are hash-tagged for it) or a managed Redis of that size |
| Inter-node connections | 2 per node pair | 58 per pod, 870 in total | Long-lived; no connection per message |
| Death detection / cleanup | 4.5–5.6 s to every client on 4 nodes | about 5 s + reap | A dead pod's 100k presences are removed with pipelined scripts, 500 per round trip (about 200 round trips) |

Where it bends first: the per-pod event fan-out of very large rooms (every member's node gets every join, leave and message, as on one node), then Redis operations per second (move to Redis Cluster before about 300k players), then per-pod CPU. Not measured here: more than 4 nodes, nodes on separate machines, Redis Cluster, a soak longer than a few minutes, and a pod dying under full load.

## Single-node regression

With clustering off the full Playwright suite (201 tests) passed after every step; on the final build two consecutive runs passed 201/201 (one earlier run on the same build failed only the external ZarinPal sandbox page, which passes on rerun and needs no server change).
