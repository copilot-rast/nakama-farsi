import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import url from 'node:url';
import zlib from 'node:zlib';

export const ROOT = path.resolve(path.dirname(url.fileURLToPath(import.meta.url)), '..');
const DIST = path.resolve(ROOT, '..', 'dist');
const INDEX = path.join(DIST, 'index.html');
const MANIFEST = path.join(ROOT, 'dist-manifest.json');
export const START = '<!-- nakama-leaderboardx:start -->';
export const END = '<!-- nakama-leaderboardx:end -->';
export const SOURCES = ['i18n.js', 'core.js', 'editors.js', 'boards.js', 'standings.js', 'periods.js', 'report.js', 'settings.js', 'player.js', 'shell.js'];

function hash(content) {
  return crypto.createHash('sha256').update(content).digest('hex').slice(0, 10);
}

export function restoreIndex(html) {
  const s = html.indexOf(START);
  const e = html.indexOf(END);
  if (s === -1 || e === -1) return html;
  return html.slice(0, s).replace(/[ \t]*$/, '') + html.slice(e + END.length).replace(/^\r?\n/, '');
}

export function injectIndex(html, cssHref, jsHref) {
  const clean = restoreIndex(html);
  const block = `    ${START}\n    <link rel="stylesheet" href="${cssHref}">\n    <script defer src="${jsHref}"></script>\n    ${END}\n`;
  const anchor = clean.indexOf('<!-- nakama-assets:start -->') !== -1 ? /[ \t]*<!-- nakama-assets:start -->/
    : clean.indexOf('<!-- nakama-campaign:start -->') !== -1 ? /[ \t]*<!-- nakama-campaign:start -->/
    : clean.indexOf('<!-- nakama-shopx:start -->') !== -1 ? /[ \t]*<!-- nakama-shopx:start -->/
    : clean.indexOf('<!-- nakama-account:start -->') !== -1 ? /[ \t]*<!-- nakama-account:start -->/
    : clean.indexOf('<!-- nakama-analytics:start -->') !== -1 ? /[ \t]*<!-- nakama-analytics:start -->/
    : clean.indexOf('<!-- nakama-devkit:start -->') !== -1 ? /[ \t]*<!-- nakama-devkit:start -->/
    : clean.indexOf('<!-- nakama-economy:start -->') !== -1 ? /[ \t]*<!-- nakama-economy:start -->/
    : clean.indexOf('<!-- nakama-l10n:start -->') !== -1 ? /[ \t]*<!-- nakama-l10n:start -->/ : /[ \t]*<script type="module"/;
  if (!anchor.test(clean)) throw new Error('dist/index.html has no console bundle script');
  return clean.replace(anchor, (m) => block + m);
}

export function bundle() {
  const parts = SOURCES.map((f) => fs.readFileSync(path.join(ROOT, 'src', f), 'utf8').trimEnd());
  const js = `(function () {\n'use strict';\n${parts.join('\n\n')}\n})();\n`;
  const css = fs.readFileSync(path.join(ROOT, 'src', 'leaderboardx.css'), 'utf8');
  return { js, css };
}

function main() {
  const check = process.argv.includes('--check');
  const { js, css } = bundle();
  const files = [];
  const out = [];
  const emit = (name, content) => {
    for (const [rel, data] of [[`static/${name}`, content], [`static/${name}.gz`, zlib.gzipSync(content, { level: 9 })]]) {
      files.push(rel);
      out.push([path.join(DIST, rel), data]);
    }
    return `/static/${name}`;
  };
  const jsHref = emit(`leaderboardx-${hash(js)}.js`, js);
  const cssHref = emit(`leaderboardx-${hash(css)}.css`, css);
  const current = fs.readFileSync(INDEX, 'utf8');
  const html = injectIndex(current, cssHref, jsHref);
  const previous = fs.existsSync(MANIFEST) ? JSON.parse(fs.readFileSync(MANIFEST, 'utf8')).files : [];
  const stale = previous.filter((f) => !files.includes(f));
  if (check) {
    const absent = out.filter(([file, data]) => !fs.existsSync(file) || (!file.endsWith('.gz') && !fs.readFileSync(file).equals(Buffer.from(data))));
    if (html !== current || absent.length || stale.length) {
      console.error('leaderboardx build: dist is out of date; run "npm run build"');
      process.exit(1);
    }
    console.log('leaderboardx: dist is up to date');
    return;
  }
  for (const f of stale) fs.rmSync(path.join(DIST, f), { force: true });
  for (const [file, data] of out) fs.writeFileSync(file, data);
  fs.writeFileSync(INDEX, html);
  fs.writeFileSync(INDEX + '.gz', zlib.gzipSync(html, { level: 9 }));
  fs.writeFileSync(MANIFEST, JSON.stringify({ note: 'Files generated in console/ui/dist by scripts/build.mjs.', files: files.sort() }, null, 2) + '\n');
  console.log(`leaderboardx: wrote ${files.length} files (${(js.length / 1024).toFixed(1)} KB js, ${(css.length / 1024).toFixed(1)} KB css); removed ${stale.length} stale`);
}

if (import.meta.url === `file://${process.argv[1]}`) main();
