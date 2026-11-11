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
