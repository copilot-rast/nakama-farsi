# Nakama Console localization (fa-IR)

Persian localization and right-to-left support for the prebuilt console in
`../dist`. The console opens in Persian by default; English is available from
the language button in the top bar or with `?lang=en`.

```sh
npm ci
npm run build    # regenerate ../dist assets and inject ../dist/index.html
npm test
npm run audit -- --list
```

Commit the regenerated `../dist` files together with source changes.
See [docs/persian-localization.md](../../../docs/persian-localization.md).
