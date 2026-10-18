// Translation completeness audit for the prebuilt console bundle.
//
//   node scripts/audit.mjs            print a summary
//   node scripts/audit.mjs --list     also list untranslated strings
//   node scripts/audit.mjs --strict   exit 1 when user-facing strings are untranslated
//   node scripts/audit.mjs --json     machine-readable report
//
// The console source is not available, so user-facing strings are recovered
// from the compiled bundle by their syntactic context: Vue text vnodes, element
// text children, UI props (label, placeholder, title, ...) and toast calls.
// Each one is run through the runtime matcher; strings that do not translate
// are reported unless audit-ignore.json classifies them (technical, brand,
// third-party, developer-only).
import * as acorn from 'acorn';
import fs from 'node:fs';
import path from 'node:path';
import vm from 'node:vm';
import { ROOT, loadCatalog, loadConfig, readJSON } from './catalog.mjs';

const DIST = path.resolve(ROOT, '..', 'dist');

const UI_PROPS = new Set([
  'label', 'title', 'subtitle', 'placeholder', 'description', 'helper', 'text', 'tooltip', 'loading-label',
  'search-placeholder', 'filter-placeholder', 'aria-label', 'alt', 'content', 'buttonLabel', 'emptyText',
  'empty-text', 'weighted-title', 'toggle-label', 'documentation-label', 'advertisingDescription', 'msg',
  'actionItem', 'hint', 'caption', 'message', 'loadingLabel', 'searchPlaceholder', 'filterPlaceholder',
]);
const UI_CALLS = new Set(['success', 'info', 'warning', 'error', 'pe']);

export function loadRuntime() {
  const sandbox = { module: { exports: {} }, Intl, URLSearchParams };
  vm.runInNewContext(fs.readFileSync(path.join(ROOT, 'src', 'runtime.js'), 'utf8'), sandbox);
  return sandbox.module.exports;
}

function bundlePath() {
  const html = fs.readFileSync(path.join(DIST, 'index.html'), 'utf8');
  const src = /<script[^>]*type="module"[^>]*src="(\/static\/index-[^"]+\.js)"/.exec(html)?.[1];
  if (!src) throw new Error('console bundle not found in dist/index.html');
  return path.join(DIST, src);
}

// extractStrings returns UI strings found in the bundle with the context that
// marked them as user-facing. Template literals become "{0}" placeholders.
export function extractStrings(code) {
  const tokens = [...acorn.tokenizer(code, { ecmaVersion: 'latest', sourceType: 'module' })];
  const found = new Map();
  const add = (text, ctx) => {
    const t = text.replace(/\s+/g, ' ').trim();
    if (!/[A-Za-z]{2}/.test(t) || t.length > 600) return;
    if (/^[a-z0-9_$.\-/:#@]+$/.test(t) && !/^[A-Z]/.test(t)) return; // identifiers, paths
    if (/<\/?(path|svg|g|line|polyline|circle|rect)\b/.test(t)) return;
    if (!found.has(t)) found.set(t, ctx);
  };
  for (let i = 0; i < tokens.length; i++) {
    const t = tokens[i];
    let value;
    let start = i;
    if (t.type.label === 'string') {
      value = t.value;
    } else if (t.type.label === '`') {
      const parts = [];
      let n = 0;
      let j = i + 1;
      for (; j < tokens.length && tokens[j].type.label !== '`'; j++) {
        const l = tokens[j].type.label;
        if (l === 'template') parts.push(tokens[j].value);
        else if (l === '${') {
          parts.push(`{${n++}}`);
          let depth = 1;
          while (depth && ++j < tokens.length) {
            const ll = tokens[j].type.label;
            if (ll === '${' || ll === '{') depth++;
            else if (ll === '}') depth--;
          }
        }
      }
      value = parts.join('');
      i = j;
    } else {
      continue;
    }
    const prev = tokens[start - 1];
    const prev2 = tokens[start - 2];
    if (prev?.type.label === ':' && prev2 && UI_PROPS.has(String(prev2.value))) add(value, `prop:${prev2.value}`);
    else if (prev?.type.label === '(' && prev2?.type.label === 'name' && UI_CALLS.has(prev2.value)) add(value, `call:${prev2.value}`);
    else if (prev?.type.label === ',' ) {
      // Third argument of an element vnode: H("div", props, "Text")
      let depth = 0;
      let k = start - 1;
      for (; k > 0 && k > start - 200; k--) {
        const l = tokens[k].type.label;
        if (l === ')' || l === ']' || l === '}') depth++;
        else if (l === '(' || l === '[' || l === '{') {
          if (depth === 0) break;
          depth--;
        }
      }
      if (tokens[k]?.type.label === '(' && tokens[k - 1]?.type.label === 'name' && tokens[k + 1]?.type.label === 'string' && /^[a-z][a-z0-9]*$/.test(tokens[k + 1].value)) {
        add(value, `element:${tokens[k + 1].value}`);
      }
    }
  }
  return found;
}

function main() {
  const args = new Set(process.argv.slice(2));
  const config = loadConfig();
  const runtime = loadRuntime();
  const ignore = readJSON('audit-ignore.json');
  const ignored = new Map();
  for (const [category, list] of Object.entries(ignore.categories)) for (const s of list) ignored.set(s, category);

  const strings = extractStrings(fs.readFileSync(bundlePath(), 'utf8'));
  const report = { bundle: path.basename(bundlePath()), locales: {} };
  for (const locale of Object.keys(config.locales)) {
    if (locale === config.sourceLocale) continue;
    const index = runtime.compile(loadCatalog(config.sourceLocale), loadCatalog(locale), locale);
    const untranslated = [];
    const byCategory = {};
    let translated = 0;
    for (const [text, ctx] of strings) {
      // Template placeholders ({0}) stand in for runtime values; probe with a sample value.
      const probe = text.replace(/\{\d+\}/g, '7');
      if (runtime.translate(index, probe) !== null || runtime.translate(index, text) !== null) {
        translated++;
      } else if (ignored.has(text)) {
        byCategory[ignored.get(text)] = (byCategory[ignored.get(text)] || 0) + 1;
      } else {
        untranslated.push({ text, ctx });
      }
    }
    report.locales[locale] = { candidates: strings.size, translated, ignored: byCategory, untranslated };
  }

  if (args.has('--json')) {
    console.log(JSON.stringify(report, null, 2));
  } else {
    console.log(`bundle ${report.bundle}`);
    for (const [locale, r] of Object.entries(report.locales)) {
      const ignoredTotal = Object.values(r.ignored).reduce((a, b) => a + b, 0);
      console.log(`${locale}: ${r.candidates} user-facing candidates, ${r.translated} translated, ${ignoredTotal} intentionally untranslated, ${r.untranslated.length} untranslated`);
      for (const [category, count] of Object.entries(r.ignored)) console.log(`  ignored ${category}: ${count}`);
      if (args.has('--list')) for (const u of r.untranslated) console.log(`  - [${u.ctx}] ${JSON.stringify(u.text)}`);
    }
  }
  if (args.has('--strict') && Object.values(report.locales).some((r) => r.untranslated.length)) process.exit(1);
}

if (import.meta.url === `file://${process.argv[1]}`) main();
