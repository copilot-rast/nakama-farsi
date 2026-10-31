# Star Catcher (ستاره‌چین)

A small web game for testing the economy, publication and base modules end to end. Plain HTML, CSS and JavaScript on a canvas; it talks to the Nakama REST API and WebSocket directly. The coin reward for a round is the Go plugin in `main.go` (`stargame_start`, `stargame_finish`).

## Play it

The plugin serves the game itself at `/stargame/` on the Nakama API port, so after a deploy it is just there:

- https://nakama.dastar.dev/stargame/
- locally: http://100.106.1.79:7450/stargame/

Served from there, the game talks to the same origin and reads the server key from `/stargame/config.json` (the `STARGAME_SERVER_KEY` runtime env, which compose and `.local/run.sh` set to the server key). The files are embedded in `stargame.so` with `go:embed`, so a change under `web/` needs a plugin rebuild.

## Run the files from a plain static server

```sh
cd samples/stargame/web
python3 -m http.server 7470
```

Open http://localhost:7470. Outside `/stargame/` it uses the Nakama API on the same host, port 7450, with server key `defaultkey`.

To play against another server, pass it once in the URL. The game remembers it in localStorage:

```
http://localhost:7470/?api=https://nakama.dastar.dev&key=<NAKAMA_SERVER_KEY>
```

`?api=reset&key=reset` goes back to the defaults.

## Shop data (automatic)

On every start the plugin creates what the game needs if it is missing, in one transaction with `ON CONFLICT DO NOTHING`, so several nodes can start together:

- languages fa, en, ar; currency `coin` (50 for new players)
- product categories `star_coin_packs`, `star_skins`; prizes `star_coins_100/500/1200`, `star_skin_gold/neon/galaxy`
- products `star_pack_small/medium/large` (Toman, ZarinPal, Zibal, Bazaar, Myket) and `star_skin_gold/neon/galaxy` (coins, in_app)
- gift codes `STAR100` (100 coins) and `HALF50` (50% off a bank payment)
- payment methods, only while a method was never saved in the console: in_app active; zarinpal and zibal active in sandbox; bazaar and myket active in sandbox with package `ir.nakama.stargame`

Rows are matched by key and never updated: names, prices or methods changed in the console stay as they are. A seeded row that an admin deletes comes back on the next start. The log shows one line, `Star Catcher seed: created …` or `nothing new`. Set the runtime env `STARGAME_SEED=false` (compose: `STARGAME_SEED=false` in the environment) to turn it off.

## Server side

- The Docker image (`build/Dockerfile.local`) ships `stargame.so` next to the other plugins.
- Bank payments return to `ECONOMY_WEB_REDIRECT_URL`; compose defaults it to `https://nakama.dastar.dev/stargame/?payment={paymentId}&success={success}`. For a static server use `http://localhost:7470/?payment={paymentId}&success={success}`.
- `seed.py` is the manual alternative. Unlike the plugin it overwrites the sample rows with its own values:

```sh
CONSOLE_URL=https://nakama-console.dastar.dev CONSOLE_USER=admin CONSOLE_PASSWORD=... python3 samples/stargame/seed.py
```
