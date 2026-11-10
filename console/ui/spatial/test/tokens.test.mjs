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

function rgba(v) {
  let m = /^#([0-9a-f]{6})$/i.exec(v);
  if (m) return [parseInt(m[1].slice(0, 2), 16), parseInt(m[1].slice(2, 4), 16), parseInt(m[1].slice(4, 6), 16), 1];
  m = /^rgb\((\d+) (\d+) (\d+)(?: \/ ([\d.]+))?\)$/.exec(v);
  assert.ok(m, 'colour ' + v);
  return [+m[1], +m[2], +m[3], m[4] === undefined ? 1 : +m[4]];
}

const over = (fg, bg) => [0, 1, 2].map((i) => fg[i] * fg[3] + bg[i] * (1 - fg[3])).concat(1);
const lin = (c) => { c /= 255; return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; };
const lum = (c) => 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]);
const ratio = (a, b) => { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); };

const light = block('html.sp');
const dark = { ...light, ...block('html.sp.dark') };
const skyLight = ['#dfe7f3', '#e9edf3', '#f4ece6', '#d5dcea', '#ccd5e6', '#c5cfe2'].map(rgba);
const skyDark = ['#0b0e1a', '#111528', '#1a1630', '#171a2e', [72, 60, 69, 1]].map((c) => (Array.isArray(c) ? c : rgba(c)));

for (const [name, t, sky] of [['light', light, skyLight], ['dark', dark, skyDark]]) {
  test(`ink text passes 4.5:1 on every glass level over the ${name} sky`, () => {
    let min = 99;
    for (const s of sky) {
      for (const level of ['glass-top', 'glass-bot', 'thick-top', 'thick-bot']) {
        const surface = over(rgba(t[level]), s);
        const fills = [surface, over(rgba(t.well), surface), over(rgba(t['well-2']), surface), over(rgba(t['capsule-hi']), surface)];
        for (const [fi, fill] of fills.entries()) {
          for (const ink of fi === 3 ? ['ink-1', 'ink-2'] : ['ink-1', 'ink-2', 'ink-3']) {
            const r = ratio(rgba(t[ink]), fill);
            min = Math.min(min, r);
            assert.ok(r >= 4.5, `${name} ${ink} on ${level}: ${r.toFixed(2)}`);
          }
          if (fi < 3) for (const role of ['pos', 'neg', 'warn']) assert.ok(ratio(rgba(t[role]), fill) >= 4.5, `${name} ${role} on ${level}: ${ratio(rgba(t[role]), fill).toFixed(2)}`);
        }
      }
    }
    assert.ok(min >= 4.5);
  });
}

test('the default accent passes: white on accent, accent ink on light glass and dark capsule', () => {
  const accent = rgba(/var\(--sp-accent, (rgb\([^)]+\))\)/.exec(light.accent)[1]);
  const ink = rgba(/var\(--sp-accent-ink, (rgb\([^)]+\))\)/.exec(light['accent-ink'])[1]);
  const inkDark = rgba(/var\(--sp-accent-ink-dark, (rgb\([^)]+\))\)/.exec(dark['accent-ink'])[1]);
  assert.ok(ratio([255, 255, 255, 1], accent) >= 4.5);
  assert.ok(ratio(ink, rgba('#dfe4ee')) >= 4.5);
  assert.ok(ratio(inkDark, rgba('#4a4d5f')) >= 4.5);
});

test('motion tokens: one easing family, four durations, exits faster, reduced motion is instant', () => {
  assert.equal(light.ease, 'cubic-bezier(.2, .8, .2, 1)');
  assert.deepEqual([light.d1, light.d2, light.d3, light.d4], ['120ms', '200ms', '280ms', '480ms']);
  assert.ok(parseInt(light['d-exit'], 10) < parseInt(light.d3, 10));
  const reduce = css.slice(css.indexOf('@media (prefers-reduced-motion: reduce)'));
  for (const d of ['d1', 'd2', 'd3', 'd4', 'd-exit']) assert.match(reduce, new RegExp(`--${d}: 1ms`));
});

test('shell strings: fa and en have the same keys and Persian copy follows the house rules', () => {
  const sandbox = {};
  vm.runInNewContext(fs.readFileSync(path.join(ROOT, 'src', 'i18n.js'), 'utf8') + ';this.M = SP_MESSAGES;', sandbox);
  const m = sandbox.M;
  assert.deepEqual(Object.keys(m.fa).sort(), Object.keys(m.en).sort());
  for (const [k, v] of Object.entries(m.fa)) {
    assert.ok(!/[—]/.test(v), k);
    assert.ok(!/(^|[\s(])(ن?می) [؀-ۿ]/.test(v), `${k}: "می" needs a ZWNJ`);
    assert.ok(!/[؀-ۿ] (ها|های|هایی|تر|ترین)([\s.،؛:)]|$)/.test(v), `${k}: suffix needs a ZWNJ`);
    assert.ok(!/[؀-ۿ],/.test(v), `${k}: Persian comma`);
  }
});
