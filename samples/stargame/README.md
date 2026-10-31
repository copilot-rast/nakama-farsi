# Star Catcher (ستاره‌چین)

A small web game for testing the economy, publication and base modules end to end. Plain HTML, CSS and JavaScript on a canvas; it talks to the Nakama REST API and WebSocket directly. The coin reward for a round is the Go plugin in `main.go` (`stargame_start`, `stargame_finish`).

## Run the game

```sh
cd samples/stargame/web
python3 -m http.server 7470
```

Open http://localhost:7470. By default it uses the Nakama API on the same host, port 7450, with server key `defaultkey`.

To play against another server, pass it once in the URL. The game remembers it in localStorage:

```
http://localhost:7470/?api=https://nakama.dastar.dev&key=<NAKAMA_SERVER_KEY>
```

`?api=reset&key=reset` goes back to the defaults.

## Server side

- The Docker image (`build/Dockerfile.local`) already ships `stargame.so` next to the other plugins.
- For bank payments to return to the game, set `ECONOMY_WEB_REDIRECT_URL=http://localhost:7470/?payment={paymentId}&success={success}` on the server.
- Seed the game's currency, categories, prizes, products and codes into a server:

```sh
CONSOLE_URL=https://nakama-console.dastar.dev CONSOLE_USER=admin CONSOLE_PASSWORD=... python3 samples/stargame/seed.py
```

Payment methods (Settings → payment gateways) must be active on that server, in sandbox mode for testing.
