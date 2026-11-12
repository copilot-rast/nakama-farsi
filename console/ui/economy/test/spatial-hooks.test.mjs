import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { ROOT, bundle } from '../scripts/build.mjs';

const brandSrc = fs.readFileSync(path.join(ROOT, 'src', 'brand.js'), 'utf8');

function pick(src, names) {
  return names.map((n) => {
    const fn = new RegExp(`^function ${n}\\([^)]*\\) \\{[\\s\\S]*?^\\}`, 'm').exec(src);
    if (fn) return fn[0];
    const v = new RegExp(`^var ${n} = .*;$`, 'm').exec(src);
    assert.ok(v, n);
    return v[0];
  }).join('\n');
}

const ctx = {};
vm.runInNewContext(pick(brandSrc, ['BRAND_SHADES', 'BRAND_HEX_RE', 'SP_INK_LIGHT_BG', 'SP_INK_DARK_BG', 'SP_INK_MIN', 'brandHex', 'hexRgb', 'brandPalette', 'luminance', 'contrastRatio', 'pickShade', 'spatialAccent', 'brandCss']) + ';this.spatialAccent = spatialAccent; this.brandCss = brandCss;', ctx);

test('the spatial accent is picked by contrast for every swatch and odd colours', () => {
  for (const hex of ['#6b59ee', '#2563eb', '#0891b2', '#059669', '#d97706', '#dc2626', '#db2777', '#475569', '#ffee00', '#22c55e', '#2457d6']) {
    const a = ctx.spatialAccent(hex);
    assert.ok(a.ratios.onAccent >= 4.5, `${hex} white on accent ${a.ratios.onAccent.toFixed(2)}`);
    assert.ok(a.ratios.ink >= 5.5, `${hex} accent ink ${a.ratios.ink.toFixed(2)}`);
    assert.ok(a.ratios.inkDark >= 5.5, `${hex} dark accent ink ${a.ratios.inkDark.toFixed(2)}`);
  }
});

test('the green brand no longer uses shade 600 at 3.81:1 for white text', () => {
  const a = ctx.spatialAccent('#059669');
  assert.equal(a.shades.accent, 600);
  assert.ok(a.ratios.onAccent > 4.5);
  assert.match(ctx.brandCss('#059669'), /--pl-accent:rgb\(5 135 95\);--pl-accent-ink:rgb\(4 113 79\);--pl-accent-ink-dark:rgb\(105 192 165\);/);
});

test('the console hook for other overlays is exposed', () => {
  const { js } = bundle();
  assert.match(js, /window\.NakamaConsole = \{/);
  for (const k of ['router', 'lang', 't', 'sections', 'badges', 'onRoute', 'projectSettings', 'brand', 'toggleTheme']) assert.match(js, new RegExp(`\\n  (get )?${k}[:(]`), k);
  assert.match(js, /r\.afterEach\(function \(to, from\) \{ fireRoute\(to, from\); \}\)/);
  assert.match(js, /nakama-console:ready/);
});
