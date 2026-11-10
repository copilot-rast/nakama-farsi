import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { CSS_SOURCES, END, JS_SOURCES, ROOT, START, bundle, injectIndex, restoreIndex } from '../scripts/build.mjs';
import * as economy from '../../economy/scripts/build.mjs';

const L10N_BUILD = path.resolve(ROOT, '..', 'l10n', 'scripts', 'build.mjs');
function splitTop(sel) {
  const out = [];
  let depth = 0;
  let cur = '';
  for (const ch of sel) {
    if (ch === '(') depth++;
    if (ch === ')') depth--;
    if (ch === ',' && !depth) { out.push(cur.trim()); cur = ''; } else cur += ch;
  }
  out.push(cur.trim());
  return out.filter(Boolean);
}

const vendor = '<!DOCTYPE html>\n<html lang="en">\n  <head>\n    <script type="module" crossorigin src="/static/index-A.js"></script>\n    <link rel="stylesheet" crossorigin href="/static/index-B.css">\n  </head>\n</html>\n';
const l10nBlock = '    <!-- nakama-l10n:start -->\n    <script src="/static/l10n-1.js"></script>\n    <!-- nakama-l10n:end -->\n';

test('the spatial block is injected once, last in the head, and removed cleanly', () => {
  const once = injectIndex(vendor, '/static/spatial-1.css', '/static/spatial-1.js');
  assert.equal(once.split(START).length, 2);
  assert.ok(once.indexOf(END) < once.indexOf('</head>'));
  assert.ok(once.indexOf(START) > once.indexOf('index-B.css'));
  assert.equal(injectIndex(once, '/static/spatial-1.css', '/static/spatial-1.js'), once);
  assert.equal(restoreIndex(once), vendor);
  assert.equal(restoreIndex(vendor), vendor);
});

test('economy, l10n and spatial agree in every build order', async () => {
  const { restoreIndex: restoreL10n } = await import(L10N_BUILD);
  const steps = {
    economy: (h) => economy.injectIndex(h, '/static/economy-1.css', '/static/economy-1.js'),
    l10n: (h) => restoreL10n(h).replace(/<html[^>]*>/, '<html lang="fa-IR" dir="rtl">').replace('    <script type="module"', l10nBlock + '    <script type="module"'),
    spatial: (h) => injectIndex(h, '/static/spatial-1.css', '/static/spatial-1.js'),
  };
  const orders = [['economy', 'l10n', 'spatial'], ['economy', 'spatial', 'l10n'], ['l10n', 'economy', 'spatial'], ['l10n', 'spatial', 'economy'], ['spatial', 'economy', 'l10n'], ['spatial', 'l10n', 'economy']];
  const results = orders.map((o) => o.reduce((h, s) => steps[s](h), vendor));
  for (const r of results) assert.equal(r, results[0]);
  const html = results[0];
  assert.ok(html.indexOf('nakama-economy:end') < html.indexOf('nakama-l10n:start'));
  assert.ok(html.indexOf('nakama-l10n:end') < html.indexOf(START));
  assert.ok(html.indexOf('index-B.css') < html.indexOf('spatial-1.css'), 'spatial stylesheet loads after the console stylesheet');
  assert.equal(results.map((r) => steps.spatial(steps.l10n(steps.economy(r)))).every((r) => r === html), true);
});

test('the bundle parses, has no comments and keeps everything under html.sp', () => {
  const { js, css } = bundle();
  assert.doesNotThrow(() => new vm.Script(js));
  for (const f of [...JS_SOURCES, ...CSS_SOURCES]) {
    const src = fs.readFileSync(path.join(ROOT, 'src', f), 'utf8');
    assert.ok(!/^\s*\/\/|\/\*/m.test(src.replace(/'[^'\n]*'/g, "''").replace(/"[^"\n]*"/g, '""')), `${f} has a comment`);
  }
  const rules = css.replace(/@keyframes[^{]+\{(?:[^{}]*\{[^}]*\})*\s*\}/g, '').match(/(^|})\s*([^{}@]+)\{/g) || [];
  const loose = rules.map((r) => r.replace(/^}?\s*/, '').replace(/\{$/, '').trim()).filter((sel) => sel && !/^(from|to|\d+%)$/.test(sel))
    .flatMap((sel) => splitTop(sel))
    .filter((s) => !/^html(\[[^\]]+\])?(\.sp|:not\(\.sp\))|^[a-z]*\.sp-|^html\[dir=/.test(s));
  assert.deepEqual(loose, []);
  assert.ok(!/(^|[^-])(margin|padding|border)-(left|right)\s*:|(^|[;{\s])(left|right)\s*:\s*[^;]*;/.test(css.replace(/\.sp-seg-ind[^}]*}/g, '').replace(/html\[dir="rtl"\] \.sp-seg-ind[^}]*}/g, '')), 'physical left/right outside the segmented indicator');
});

test('turning spatial off leaves the page untouched', () => {
  const { js } = bundle();
  const classes = [];
  const store = { 'nakama.spatial': '0' };
  const sandbox = {
    location: { search: '', hash: '#/' },
    localStorage: { getItem: (k) => store[k] ?? null, setItem: (k, v) => { store[k] = v; }, removeItem: (k) => { delete store[k]; } },
    document: { documentElement: { classList: { add: (c) => classes.push(c) } } },
    window: {},
  };
  vm.runInNewContext(js, sandbox);
  assert.deepEqual(classes, []);
  sandbox.location.search = '?spatial=1';
  try { vm.runInNewContext(js, sandbox); } catch (e) { assert.ok(e instanceof Error); }
  assert.deepEqual(classes, ['sp']);
  assert.equal(store['nakama.spatial'], undefined);
});

test('committed dist output matches the sources', () => {
  execFileSync(process.execPath, [path.join(ROOT, 'scripts', 'build.mjs'), '--check'], { stdio: 'pipe' });
});
