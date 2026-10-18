// Builds the console localization layer into console/ui/dist.
//
//   node scripts/build.mjs           build and inject into dist/index.html
//   node scripts/build.mjs --check   validate catalogs and verify dist is up to date
//
// Generated files are content-hashed and written next to the console bundle in
// dist/static (with .gz siblings, which the server requires for .js/.css under
// /static). The vendor bundle itself is never modified; index.html only gains a
// marked block that loads the runtime, plus a switch to the logical stylesheet.
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { ROOT, loadCatalog, loadConfig, stripMeta, validateCatalogs } from './catalog.mjs';
import { toLogical } from './logical-css.mjs';

const DIST = path.resolve(ROOT, '..', 'dist');
const STATIC = path.join(DIST, 'static');
const INDEX = path.join(DIST, 'index.html');
const MANIFEST = path.join(ROOT, 'dist-manifest.json');
const START = '<!-- nakama-l10n:start -->';
const END = '<!-- nakama-l10n:end -->';
const CHECK = process.argv.includes('--check');
const ALLOW_MISSING = process.argv.includes('--allow-missing');

function hash(content) {
  return crypto.createHash('sha256').update(content).digest('hex').slice(0, 10);
}

function fail(message) {
  console.error(`l10n build: ${message}`);
  process.exit(1);
}

// restoreIndex removes a previous injection so the build is idempotent and
// works again after the upstream console bundle (and index.html) is updated.
export function restoreIndex(html) {
  let out = html;
  const s = out.indexOf(START);
  const e = out.indexOf(END);
  if (s !== -1 && e !== -1) out = out.slice(0, s).replace(/[ \t]*$/, '') + out.slice(e + END.length).replace(/^\r?\n/, '');
  out = out.replace(/<link([^>]*?)href="[^"]*"([^>]*?)\s*data-l10n-source="([^"]*)"([^>]*)>/, '<link$1href="$3"$2$4>');
  out = out.replace(/<html[^>]*>/, '<html lang="en">');
  return out;
}

// Icons are registered in the bundle as `name:'<svg markup>'` or `"name":{icon:'...'}`.
export function extractIcons(bundle, names) {
  const icons = {};
  for (const name of names) {
    const key = /^[a-z][a-z0-9]*$/.test(name) ? `(?:"${name}"|\\b${name})` : `"${name}"`;
    const m = new RegExp(`${key}:(?:\\{icon:)?'(<[^']*)'`).exec(bundle);
    if (m) icons[name] = m[1];
  }
  return icons;
}

function writeAsset(name, content, files, { gzip = false } = {}) {
  const file = path.join(STATIC, name);
  files.push(path.relative(DIST, file));
  if (!CHECK) fs.writeFileSync(file, content);
  if (gzip) {
    files.push(path.relative(DIST, file) + '.gz');
    if (!CHECK) fs.writeFileSync(file + '.gz', zlib.gzipSync(content, { level: 9 }));
  }
  return `/static/${name}`;
}

function main() {
  const config = loadConfig();
  const source = loadCatalog(config.sourceLocale);
  const catalogs = {};
  let problems = 0;
  for (const locale of Object.keys(config.locales)) {
    if (locale === config.sourceLocale) continue;
    const catalog = loadCatalog(locale);
    const { errors, missing, keys } = validateCatalogs(source, catalog, { targetLocale: locale });
    errors.forEach((e) => console.error(`  error: ${e}`));
    if (missing.length) {
      console[ALLOW_MISSING ? 'warn' : 'error'](`  ${locale}: ${missing.length} missing translation(s): ${missing.slice(0, 20).join(', ')}`);
    }
    problems += errors.length + (ALLOW_MISSING ? 0 : missing.length);
    console.log(`${locale}: ${keys - missing.length}/${keys} messages translated`);
    catalogs[locale] = stripMeta(catalog);
  }
  if (problems) fail(`${problems} catalog problem(s)`);

  const original = restoreIndex(fs.readFileSync(INDEX, 'utf8'));
  const cssHref = /<link[^>]*rel="stylesheet"[^>]*href="(\/static\/index-[^"]+\.css)"/.exec(original)?.[1];
  const jsSrc = /<script[^>]*type="module"[^>]*src="(\/static\/index-[^"]+\.js)"/.exec(original)?.[1];
  if (!cssHref || !jsSrc) fail('could not find the console stylesheet/script in dist/index.html');

  const files = [];
  const bundle = fs.readFileSync(path.join(DIST, jsSrc), 'utf8');
  const icons = extractIcons(bundle, config.mirrorIcons);
  const lostIcons = config.mirrorIcons.filter((n) => !icons[n]);
  if (lostIcons.length) console.warn(`  warning: icons not found in bundle: ${lostIcons.join(', ')}`);

  // 1. Direction-agnostic copy of the console stylesheet.
  const { css: logical, stats } = toLogical(fs.readFileSync(path.join(DIST, cssHref), 'utf8'));
  const cssBase = path.basename(cssHref, '.css');
  const logicalHref = writeAsset(`${cssBase}.logical-${hash(logical)}.css`, logical, files, { gzip: true });
  console.log(`logical css: ${stats.converted} declarations converted, ${stats.rtlOverrides} rtl overrides, ${stats.excludedRules} editor rules untouched`);

  // 2. Fonts and the localization stylesheet.
  let l10nCss = fs.readFileSync(path.join(ROOT, 'src', 'l10n.css'), 'utf8');
  const fontHrefs = {};
  for (const [id, file] of Object.entries(config.fonts)) {
    const content = fs.readFileSync(path.join(ROOT, file));
    const name = `${path.basename(file, '.woff2').toLowerCase()}-${hash(content)}.woff2`;
    fontHrefs[id] = writeAsset(name, content, files);
    l10nCss = l10nCss.replaceAll(`__FONT_${id}__`, fontHrefs[id]);
  }
  if (/__FONT_\w+__/.test(l10nCss)) fail('unresolved font placeholder in src/l10n.css');
  writeAsset('vazirmatn-OFL.txt', fs.readFileSync(path.join(ROOT, 'fonts', 'OFL.txt')), files);
  const cssOut = `/* Nakama Console localization. Vazirmatn font: SIL Open Font License 1.1, see /static/vazirmatn-OFL.txt */\n${l10nCss}`;
  const l10nCssHref = writeAsset(`l10n-${hash(cssOut)}.css`, cssOut, files, { gzip: true });

  // 3. Runtime with the catalogs embedded (loaded synchronously before the app).
  const locales = {};
  for (const [code, info] of Object.entries(config.locales)) locales[code] = info;
  const build = {
    version: hash(JSON.stringify([source, catalogs])),
    defaultLocale: config.defaultLocale,
    locales,
    source: stripMeta(source),
    catalogs,
    mirrorIcons: Object.values(icons),
  };
  const runtime = fs.readFileSync(path.join(ROOT, 'src', 'runtime.js'), 'utf8');
  const js = `/* Nakama Console localization runtime. Generated by console/ui/l10n/scripts/build.mjs; do not edit. */\nvar __NAKAMA_L10N_BUILD__ = ${JSON.stringify(build)};\n${runtime}`;
  const jsHref = writeAsset(`l10n-${hash(js)}.js`, js, files, { gzip: true });

  // 4. index.html: default document language/direction, logical stylesheet, runtime.
  const info = config.locales[config.defaultLocale];
  const preload = (config.preloadFonts || [])
    .map((id) => `    <link rel="preload" href="${fontHrefs[id]}" as="font" type="font/woff2" crossorigin>\n`)
    .join('');
  const block =
    `    ${START}\n${preload}` +
    `    <link rel="stylesheet" href="${l10nCssHref}">\n` +
    `    <script src="${jsHref}"></script>\n` +
    `    ${END}\n`;
  let html = original.replace(/<html[^>]*>/, `<html lang="${info.lang}" dir="${info.dir}">`);
  html = html.replace(/(<link[^>]*rel="stylesheet"[^>]*)href="(\/static\/index-[^"]+\.css)"/, `$1href="${logicalHref}" data-l10n-source="$2"`);
  html = html.replace(/([ \t]*)<script type="module"/, (m) => block + m);
  const indexChanged = fs.readFileSync(INDEX, 'utf8') !== html;

  const previous = fs.existsSync(MANIFEST) ? JSON.parse(fs.readFileSync(MANIFEST, 'utf8')).files : [];
  const stale = previous.filter((f) => !files.includes(f));
  if (CHECK) {
    const absent = files.filter((f) => !fs.existsSync(path.join(DIST, f)));
    if (indexChanged || absent.length || stale.length) {
      fail(`dist is out of date (index changed: ${indexChanged}, missing: ${absent.join(', ') || 'none'}); run "npm run build"`);
    }
    console.log('dist is up to date');
    return;
  }
  for (const f of stale) fs.rmSync(path.join(DIST, f), { force: true });
  fs.writeFileSync(INDEX, html);
  fs.writeFileSync(INDEX + '.gz', zlib.gzipSync(html, { level: 9 }));
  fs.writeFileSync(MANIFEST, JSON.stringify({ note: 'Files generated in console/ui/dist by scripts/build.mjs.', files: files.sort() }, null, 2) + '\n');
  console.log(`wrote ${files.length} files; removed ${stale.length} stale`);
}

if (import.meta.url === `file://${process.argv[1]}`) main();
