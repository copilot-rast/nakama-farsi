import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { ROOT } from '../scripts/build.mjs';

const css = fs.readFileSync(path.join(ROOT, 'src', 'tokens.css'), 'utf8');

function block(selector) {
  const i = css.indexOf(selector + ' {');
  assert.ok(i !== -1, selector);
  const body = css.slice(i, css.indexOf('}', i));
  const out = {};
  for (const m of body.matchAll(/--([a-z0-9-]+):\s*([^;]+);/g)) out[m[1]] = m[2].trim();
  return out;
}

function oklch(v) {
  const m = /^oklch\(([\d.]+) ([\d.]+) ([\d.]+)(?: \/ \.?([\d.]+))?\)$/.exec(v.trim());
  assert.ok(m, 'colour ' + v);
  const L = +m[1], C = +m[2], h = (+m[3] * Math.PI) / 180;
  const a = C * Math.cos(h), b = C * Math.sin(h);
  const l_ = L + 0.3963377774 * a + 0.2158037573 * b;
  const m_ = L - 0.1055613458 * a - 0.0638541728 * b;
  const s_ = L - 0.0894841775 * a - 1.291485548 * b;
  const l = l_ ** 3, mm = m_ ** 3, s = s_ ** 3;
  const lin = [4.0767416621 * l - 3.3077115913 * mm + 0.2309699292 * s, -1.2684380046 * l + 2.6097574011 * mm - 0.3413193965 * s, -0.0041960863 * l - 0.7034186147 * mm + 1.707614701 * s];
  const enc = (x) => { x = Math.min(1, Math.max(0, x)); return 255 * (x <= 0.0031308 ? 12.92 * x : 1.055 * x ** (1 / 2.4) - 0.055); };
  const alpha = m[4] === undefined ? 1 : parseFloat((v.includes('/ .') ? '.' : '') + m[4]);
  return [...lin.map(enc), alpha];
}

const over = (fg, bg) => [0, 1, 2].map((i) => fg[i] * fg[3] + bg[i] * (1 - fg[3])).concat(1);
const lin = (c) => { c /= 255; return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; };
const lum = (c) => 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]);
const ratio = (a, b) => { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); };

const light = block('html.pl');
const dark = { ...light, ...block('html.pl.dark') };

for (const [name, t] of [['light', light], ['dark', dark]]) {
  test(`ink text passes 4.5:1 on every ${name} surface`, () => {
    for (const bgName of ['base', 'surface', 'surface-2', 'well', 'well-2']) {
      const bg = oklch(t[bgName]);
      for (const ink of ['ink-1', 'ink-2', 'ink-3']) {
        const r = ratio(oklch(t[ink]), bg);
        assert.ok(r >= 4.5, `${name} ${ink} on ${bgName}: ${r.toFixed(2)}`);
      }
      for (const st of ['pos', 'neg', 'warn']) {
        const r = ratio(oklch(t[st]), bg);
        assert.ok(r >= 4.5, `${name} ${st} on ${bgName}: ${r.toFixed(2)}`);
      }
    }
  });
}

test('the default accent passes: white on accent, accent ink on surfaces in both themes', () => {
  const accent = oklch(/var\(--pl-accent, (oklch\([^)]+\))\)/.exec(light.accent)[1]);
  const ink = oklch(/var\(--pl-accent-ink, (oklch\([^)]+\))\)/.exec(light['accent-ink'])[1]);
  const inkDark = oklch(/var\(--pl-accent-ink-dark, (oklch\([^)]+\))\)/.exec(dark['accent-ink'])[1]);
  assert.ok(ratio([255, 255, 255, 1], accent) >= 4.5);
  assert.ok(ratio(ink, oklch(light.surface)) >= 4.5);
  assert.ok(ratio(ink, oklch(light.base)) >= 4.5);
  assert.ok(ratio(inkDark, oklch(dark.surface)) >= 4.5);
  assert.ok(ratio(inkDark, oklch(dark['well-2'])) >= 4.5);
});

test('motion tokens: one easing, short durations, exits faster, reduced motion is instant', () => {
  assert.equal(light.ease, 'cubic-bezier(0, 0, .2, 1)');
  assert.ok([light.d1, light.d2, light.d3, light.d4].every((d) => parseInt(d, 10) <= 300));
  assert.ok(parseInt(light['d-exit'], 10) < parseInt(light.d3, 10));
  const reduce = css.slice(css.indexOf('@media (prefers-reduced-motion: reduce)'));
  for (const d of ['d1', 'd2', 'd3', 'd4', 'd-exit']) assert.match(reduce, new RegExp(`--${d}: 1ms`));
});

test('shell strings: fa and en have the same keys and Persian copy follows the house rules', () => {
  const sandbox = {};
  vm.runInNewContext(fs.readFileSync(path.join(ROOT, 'src', 'i18n.js'), 'utf8') + ';this.M = PL_MESSAGES;', sandbox);
  const m = sandbox.M;
  assert.deepEqual(Object.keys(m.fa).sort(), Object.keys(m.en).sort());
  for (const [k, v] of Object.entries(m.fa)) {
    assert.ok(!/[—]/.test(v), k);
    assert.ok(!/(^|[\s(])(ن?می) [؀-ۿ]/.test(v), `${k}: "می" needs a ZWNJ`);
    assert.ok(!/[؀-ۿ] (ها|های|هایی|تر|ترین)([\s.،؛:)]|$)/.test(v), `${k}: suffix needs a ZWNJ`);
    assert.ok(!/[؀-ۿ],/.test(v), `${k}: Persian comma`);
  }
});
