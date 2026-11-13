import http from 'node:http';

const PORT = Number(process.env.E2E_FAKE_GITHUB_PORT || 7489);
const HOST = process.env.E2E_FAKE_GITHUB_HOST || '127.0.0.1';

const sha = (n) => n.toString(16).padStart(4, '0').repeat(10);

const COMMITS = [
  { message: 'Add a session cache warm-up flag', author: 'zyro', date: '2026-11-12T18:05:00Z' },
  { message: 'Fix matchmaker property filters with nested arrays (#2551)', author: 'mofirouz', date: '2026-11-13T07:10:00Z' },
  { message: 'feat(console)!: drop the legacy storage editor\n\nBREAKING CHANGE: the old editor route is removed', author: 'novabyte', date: '2026-11-13T09:20:00Z' },
  { message: 'Release v3.43.0', author: 'zyro', date: '2026-11-13T10:00:00Z' },
];

const FILES = [
  { filename: 'migrate/sql/20261201000000-session-cache.sql', status: 'added', additions: 14, deletions: 0 },
  { filename: 'server/session_cache.go', status: 'modified', additions: 42, deletions: 8, patch: '@@ -10,3 +10,4 @@\n-\tttl := 60\n+\tttl := cfg.Session.CacheWarmup\n+\tif v := os.Getenv("NAKAMA_SESSION_WARMUP"); v != "" {' },
  { filename: 'server/matchmaker.go', status: 'modified', additions: 12, deletions: 3 },
  { filename: 'server/matchmaker_test.go', status: 'modified', additions: 30, deletions: 0 },
  { filename: 'console/ui/src/app/storage/storage.component.ts', status: 'removed', additions: 0, deletions: 120 },
  { filename: 'build/Dockerfile', status: 'modified', additions: 1, deletions: 1, patch: '@@ -1 +1 @@\n-FROM golang:1.27.3-bookworm AS builder\n+FROM golang:1.27.4-bookworm AS builder' },
  { filename: 'go.mod', status: 'modified', additions: 2, deletions: 2, patch: '@@ -1,4 +1,4 @@\n-go 1.27.3\n+go 1.27.4\n-\tgithub.com/jackc/pgx/v5 v5.7.4\n+\tgithub.com/jackc/pgx/v5 v5.7.5' },
  { filename: 'vendor/github.com/jackc/pgx/v5/conn.go', status: 'modified', additions: 6, deletions: 2 },
  { filename: 'vendor/modules.txt', status: 'modified', additions: 1, deletions: 1 },
  { filename: 'CHANGELOG.md', status: 'modified', additions: 9, deletions: 1 },
];

const RELEASES = [
  { tag_name: 'v3.43.0', name: 'v3.43.0', published_at: '2026-11-13T10:05:00Z', body: '### Added\n- Session cache warm-up flag.\n\n### Changed\n- Build with Go 1.27.4.\n\n### Removed\n- The legacy storage editor in the console.\n\n### Fixed\n- Matchmaker property filters with nested arrays.' },
  { tag_name: 'v3.42.0', name: 'v3.42.0', published_at: '2026-11-03T12:00:00Z', body: '### Added\n- Hiro Streaks support in the console.' },
];

const fresh = () => ({ ahead: 4, status: 0, retryAfter: 0, version: 1 });
let scenario = fresh();
let calls = { compare: 0, releases: 0, tags: 0, notModified: 0, auth: 0 };

function json(res, status, body, headers = {}) {
  res.writeHead(status, { 'Content-Type': 'application/json', ...headers });
  res.end(body === undefined ? '' : JSON.stringify(body));
}

function compareBody(owner, repo) {
  const n = Math.max(0, Math.min(COMMITS.length, scenario.ahead));
  const commits = COMMITS.slice(0, n).map((c, i) => ({
    sha: sha(i + 1),
    html_url: `https://github.com/${owner}/${repo}/commit/${sha(i + 1)}`,
    commit: { message: c.message, author: { name: c.author, date: c.date }, committer: { date: c.date } },
    author: { login: c.author },
    parents: [{ sha: i ? sha(i) : 'base' }],
  }));
  return { status: n ? 'ahead' : 'identical', ahead_by: n, behind_by: 0, total_commits: n, commits, files: n ? FILES : [] };
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://${req.headers.host}`);
  if (url.pathname === '/__fake/scenario') {
    if (req.method === 'POST') {
      let raw = '';
      req.on('data', (c) => { raw += c; });
      req.on('end', () => {
        let body = {};
        try { body = raw ? JSON.parse(raw) : {}; } catch { body = {}; }
        scenario = body.reset ? fresh() : { ...scenario, ...body, version: scenario.version + 1 };
        if (body.reset) calls = { compare: 0, releases: 0, tags: 0, notModified: 0, auth: 0 };
        json(res, 200, { scenario, calls });
      });
      return;
    }
    json(res, 200, { scenario, calls });
    return;
  }
  const m = /^\/repos\/([^/]+)\/([^/]+)\/(compare\/.+|releases|tags)$/.exec(url.pathname);
  if (!m || req.method !== 'GET') { json(res, 404, { message: 'Not Found' }); return; }
  const [, owner, repo, what] = m;
  const kind = what.startsWith('compare') ? 'compare' : what;
  calls[kind] += 1;
  if (req.headers.authorization) calls.auth += 1;
  const rate = { 'X-RateLimit-Limit': '60', 'X-RateLimit-Remaining': String(Math.max(0, 60 - calls.compare * 3)), 'X-RateLimit-Reset': String(Math.floor(Date.now() / 1000) + 3600) };
  if (scenario.status === 429 || scenario.status === 403) {
    json(res, scenario.status, { message: 'API rate limit exceeded' }, { ...rate, 'X-RateLimit-Remaining': '0', ...(scenario.retryAfter ? { 'Retry-After': String(scenario.retryAfter) } : {}) });
    return;
  }
  if (scenario.status) { json(res, scenario.status, { message: 'Server Error' }, rate); return; }
  const etag = `"${kind}-${owner}-${repo}-${scenario.version}"`;
  if (req.headers['if-none-match'] === etag) {
    calls.notModified += 1;
    res.writeHead(304, { ETag: etag, ...rate });
    res.end();
    return;
  }
  let body;
  if (kind === 'compare') body = compareBody(owner, repo);
  else if (kind === 'releases') body = RELEASES.map((r) => ({ ...r, html_url: `https://github.com/${owner}/${repo}/releases/tag/${r.tag_name}`, draft: false, prerelease: false }));
  else body = [{ name: 'v3.43.0', commit: { sha: sha(4) } }, { name: 'v3.42.0', commit: { sha: 'base' } }];
  json(res, 200, body, { ETag: etag, ...rate });
});

server.listen(PORT, HOST, () => {
  process.stdout.write(`fake github on http://${HOST}:${PORT}\n`);
});
