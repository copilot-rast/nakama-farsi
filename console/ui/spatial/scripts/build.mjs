import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import url from 'node:url';
import zlib from 'node:zlib';

export const ROOT = path.resolve(path.dirname(url.fileURLToPath(import.meta.url)), '..');
export const DIST = path.resolve(ROOT, '..', 'dist');
const STATIC = path.join(DIST, 'static');
const INDEX = path.join(DIST, 'index.html');
const MANIFEST = path.join(ROOT, 'dist-manifest.json');
export const SELECTORS = path.join(ROOT, 'selectors.json');
export const START = '<!-- nakama-spatial:start -->';
export const END = '<!-- nakama-spatial:end -->';
export const JS_SOURCES = ['i18n.js', 'core.js'];
export const CSS_SOURCES = ['tokens.css'];

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
  const block = `    ${START}\n    <link rel="stylesheet" href="${cssHref}">\n    <script src="${jsHref}"></script>\n    ${END}\n`;
  const anchor = /[ \t]*<\/head>/;
  if (!anchor.test(clean)) throw new Error('dist/index.html has no </head>');
  return clean.replace(anchor, (m) => block + m);
}

export function bundle() {
  const parts = JS_SOURCES.map((f) => fs.readFileSync(path.join(ROOT, 'src', f), 'utf8').trimEnd());
  const js = `(function () {\n'use strict';\n${parts.join('\n\n')}\n})();\n`;
  const css = CSS_SOURCES.map((f) => fs.readFileSync(path.join(ROOT, 'src', f), 'utf8').trimEnd()).join('\n') + '\n';
  return { js, css };
}

export function cssClasses(css) {
  const out = new Set();
  for (const m of css.matchAll(/\.((?:\\.|[A-Za-z_])(?:\\.|[A-Za-z0-9_-])*)/g)) out.add(m[1].replace(/\\(.)/g, '$1'));
  return out;
}

export function nativeSources(dir = STATIC) {
  const names = fs.readdirSync(dir);
  const css = names.filter((f) => /^index-[^.]+\.css$/.test(f));
  const js = names.filter((f) => /^index-[^.]+\.js$/.test(f));
  if (!css.length || !js.length) throw new Error('native console bundle not found in ' + dir);
  return {
    css: css.map((f) => fs.readFileSync(path.join(dir, f), 'utf8')).join('\n'),
    js: js.map((f) => fs.readFileSync(path.join(dir, f), 'utf8')).join('\n'),
  };
}

export function loadSelectors(file = SELECTORS) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function inJs(js, cls) {
  const m = /^x-([a-z-]+)$/.exec(cls);
  if (m) {
    const pascal = m[1].split('-').map((w) => w[0].toUpperCase() + w.slice(1)).join('');
    if (js.includes('`X' + pascal + '`') || js.includes('av(`' + pascal + '`')) return true;
  }
  const esc = cls.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp('[\\s"\'`]' + esc + '[\\s"\'`]').test(js);
}

export function missingSelectors(native, manifest) {
  const known = cssClasses(native.css);
  const out = [];
  for (const entry of manifest.selectors) {
    for (const c of entry.classes || []) if (!known.has(c) && !inJs(native.js, c)) out.push(`${entry.id}: .${c}`);
  }
  return out;
}

export function unlistedClasses(spatialCss, manifest) {
  const listed = new Set(manifest.selectors.flatMap((e) => e.classes || []));
  const own = new RegExp(manifest.own);
  return [...cssClasses(spatialCss)].filter((c) => !own.test(c) && !listed.has(c)).sort();
}

function main() {
  const check = process.argv.includes('--check');
  const { js, css } = bundle();
  const manifest = loadSelectors();
  const native = nativeSources();
  const missing = missingSelectors(native, manifest);
  const unlisted = unlistedClasses(css, manifest);
  if (missing.length || unlisted.length) {
    if (missing.length) console.error(`spatial: native classes gone from the console bundle:\n  ${missing.join('\n  ')}`);
    if (unlisted.length) console.error(`spatial: native classes used in src but not listed in selectors.json:\n  ${unlisted.join('\n  ')}`);
    process.exit(1);
  }
  const files = [];
  const out = [];
  const emit = (name, content) => {
    for (const [rel, data] of [[`static/${name}`, content], [`static/${name}.gz`, zlib.gzipSync(content, { level: 9 })]]) {
      files.push(rel);
      out.push([path.join(DIST, rel), data]);
    }
    return `/static/${name}`;
  };
  const jsHref = emit(`spatial-${hash(js)}.js`, js);
  const cssHref = emit(`spatial-${hash(css)}.css`, css);
  const current = fs.readFileSync(INDEX, 'utf8');
  const html = injectIndex(current, cssHref, jsHref);
  const previous = fs.existsSync(MANIFEST) ? JSON.parse(fs.readFileSync(MANIFEST, 'utf8')).files : [];
  const stale = previous.filter((f) => !files.includes(f));

  if (check) {
    const absent = out.filter(([file, data]) => !fs.existsSync(file) || (!file.endsWith('.gz') && !fs.readFileSync(file).equals(Buffer.from(data))));
    const gzStale = fs.existsSync(INDEX + '.gz') && zlib.gunzipSync(fs.readFileSync(INDEX + '.gz')).toString('utf8') !== current;
    if (html !== current || absent.length || stale.length || gzStale) {
      console.error(`spatial build: dist is out of date (index changed: ${html !== current}, missing: ${absent.map(([f]) => path.relative(DIST, f)).join(', ') || 'none'}, stale: ${stale.join(', ') || 'none'}, index.html.gz stale: ${gzStale}); run "npm run build"`);
      process.exit(1);
    }
    console.log(`spatial: dist is up to date; ${manifest.selectors.length} native selectors present`);
    return;
  }
  for (const f of stale) fs.rmSync(path.join(DIST, f), { force: true });
  for (const [file, data] of out) fs.writeFileSync(file, data);
  fs.writeFileSync(INDEX, html);
  fs.writeFileSync(INDEX + '.gz', zlib.gzipSync(html, { level: 9 }));
  fs.writeFileSync(MANIFEST, JSON.stringify({ note: 'Files generated in console/ui/dist by scripts/build.mjs.', files: files.sort() }, null, 2) + '\n');
  console.log(`spatial: wrote ${files.length} files (${(js.length / 1024).toFixed(1)} KB js, ${(css.length / 1024).toFixed(1)} KB css); removed ${stale.length} stale`);
}

if (import.meta.url === `file://${process.argv[1]}`) main();
