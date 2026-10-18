import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import test from 'node:test';
import { extractIcons, restoreIndex } from '../scripts/build.mjs';
import { negate, negateTransform, toLogical } from '../scripts/logical-css.mjs';
import { ROOT } from '../scripts/catalog.mjs';

test('physical properties become logical ones', () => {
  const { css } = toLogical('.a{margin-left:1rem;padding-right:2px;left:0;text-align:right;border-top-left-radius:4px;float:left}');
  assert.equal(css, '.a{margin-inline-start:1rem;padding-inline-end:2px;inset-inline-start:0;text-align:end;border-start-start-radius:4px;float:inline-start}');
});

test('asymmetric shorthands are split; symmetric ones are kept', () => {
  assert.equal(toLogical('.a{margin:1px 2px 3px 4px}').css, '.a{margin-block:1px 3px;margin-inline:4px 2px}');
  assert.equal(toLogical('.a{padding:1px 2px 3px 2px}').css, '.a{padding:1px 2px 3px 2px}');
  assert.equal(toLogical('.a{border-radius:4px 0 0 4px}').css, '.a{border-start-start-radius:4px;border-start-end-radius:0;border-end-end-radius:0;border-end-start-radius:4px}');
});

test('horizontal translations get a negated RTL override', () => {
  const { css } = toLogical('.x{--tw-translate-x: 1rem}@media (min-width:1px){.y{transform:translateX(-100%)}}');
  assert.match(css, /html\[dir="rtl"\] \.x\{--tw-translate-x: ?-1rem\}/);
  assert.match(css, /@media \(min-width:1px\)\{\.y\{transform:translateX\(-100%\)\}html\[dir="rtl"\] \.y\{transform: ?translateX\(100%\)\}\}/);
  assert.equal(negate('calc(var(--a) * 2)'), 'calc(-1 * calc(var(--a) * 2))');
  assert.equal(negate('0'), '0');
  assert.equal(negateTransform('rotate(45deg)'), null);
});

test('editor rules and keyframes are left untouched', () => {
  const input = '.monaco-editor .x{margin-left:3px}@keyframes s{from{left:0}to{left:10px}}';
  assert.equal(toLogical(input).css, input);
});

test('index injection can be removed again (idempotent rebuilds)', () => {
  const original = '<!DOCTYPE html>\n<html lang="en">\n  <head>\n    <script type="module" crossorigin src="/static/index-A.js"></script>\n    <link rel="stylesheet" crossorigin href="/static/index-B.css">\n  </head>\n</html>\n';
  const injected = original
    .replace('<html lang="en">', '<html lang="fa-IR" dir="rtl">')
    .replace('href="/static/index-B.css"', 'href="/static/index-B.logical-1.css" data-l10n-source="/static/index-B.css"')
    .replace('    <script type="module"', '    <!-- nakama-l10n:start -->\n    <script src="/static/l10n-1.js"></script>\n    <!-- nakama-l10n:end -->\n    <script type="module"');
  assert.equal(restoreIndex(injected), original);
  assert.equal(restoreIndex(original), original);
});

test('icons are extracted from the bundle registry by name', () => {
  const bundle = `x={"chevron-left":'<polyline points="15 18 9 12 15 6"></polyline>',"first-page":{icon:'<path d="M7 18"/>',filled:!0},back:'<path d="M1"/>'}`;
  assert.deepEqual(extractIcons(bundle, ['chevron-left', 'first-page', 'back', 'missing']), {
    'chevron-left': '<polyline points="15 18 9 12 15 6"></polyline>',
    'first-page': '<path d="M7 18"/>',
    back: '<path d="M1"/>',
  });
});

test('committed dist output matches the sources', () => {
  execFileSync(process.execPath, [path.join(ROOT, 'scripts', 'build.mjs'), '--check'], { stdio: 'pipe' });
});

test('translation audit finds no untranslated user-facing strings', () => {
  execFileSync(process.execPath, [path.join(ROOT, 'scripts', 'audit.mjs'), '--strict'], { stdio: 'pipe' });
});
