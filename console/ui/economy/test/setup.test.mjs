import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { ROOT } from '../scripts/build.mjs';

function brandHelpers() {
  const src = fs.readFileSync(path.join(ROOT, 'src', 'brand.js'), 'utf8');
  const pick = (name) => {
    const start = src.indexOf('function ' + name + '(');
    assert.ok(start !== -1, name);
    let depth = 0;
    for (let i = src.indexOf('{', start); i < src.length; i++) {
      if (src[i] === '{') depth++;
      if (src[i] === '}' && --depth === 0) return src.slice(start, i + 1);
    }
    throw new Error(name);
  };
  const consts = src.split('\n').filter((l) => /^var (BRAND_SHADES|BRAND_HEX_RE|BRAND_DEFAULT_SHADES|BRAND_RGB_RE|SP_INK_LIGHT_BG|SP_INK_DARK_BG|SP_INK_MIN) /.test(l)).join('\n');
  const names = ['brandHex', 'hexRgb', 'brandPalette', 'luminance', 'contrastRatio', 'brandContrast', 'contrastLevel', 'pickShade', 'spatialAccent', 'brandCss'];
  const sandbox = {};
  vm.runInNewContext(consts + '\n' + names.map(pick).join('\n') + ';this.out = {' + names.join(',') + '};', sandbox);
  return sandbox.out;
}

test('brand colors: hex parsing, palette and contrast', () => {
  const b = brandHelpers();
  assert.equal(b.brandHex('#0EA'), '#00eeaa');
  assert.equal(b.brandHex('6b59ee'), '#6b59ee');
  assert.equal(b.brandHex('purple'), '');
  const p = b.brandPalette('#6b59ee');
  assert.deepEqual(Array.from(p[500]), [107, 89, 238]);
  assert.ok(p[50][0] > p[500][0] && p[900][0] < p[500][0]);
  assert.equal(Math.round(b.contrastRatio([0, 0, 0], [255, 255, 255])), 21);
  assert.equal(b.contrastLevel(b.brandContrast('#ffee00').light), 'poor');
  assert.equal(b.contrastLevel(b.brandContrast('#1e3a8a').light), 'good');
  assert.match(b.brandCss('#059669'), /--primary-color-500:5 150 105;/);
});

test('the default accent palette maps to every shade the native bundle hardcodes', () => {
  const src = fs.readFileSync(path.join(ROOT, 'src', 'brand.js'), 'utf8');
  for (const v of ['244,244,253', '107,89,238', '102,84,226', '130,117,240', '69,58,154', '48,40,106']) assert.ok(src.includes("'" + v + "'"), v);
  for (const f of ['ui.js', 'shell.js', 'economy.css']) {
    const s = fs.readFileSync(path.join(ROOT, 'src', f), 'utf8');
    assert.ok(!/rgb\(\s*(107|102)[\s,]+(89|84)[\s,]+(238|226)\s*\)/.test(s), f + ' still hardcodes the accent');
  }
});

test('every setup step, status and channel has fa and en copy', () => {
  const sandbox = {};
  vm.runInNewContext(fs.readFileSync(path.join(ROOT, 'src', 'i18n.js'), 'utf8') + ';this.M = MESSAGES;', sandbox);
  const m = sandbox.M;
  const keys = ['tab.brand', 'tab.setup', 'desc.brand', 'desc.setup', 'brand.signIn'];
  for (const k of ['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review']) keys.push('setup.step.' + k, 'setup.step.' + k + '.title', 'setup.step.' + k + '.desc');
  for (const s of ['done', 'skipped', 'pending']) keys.push('setup.st.' + s);
  for (const s of ['pending', 'in_progress', 'completed', 'skipped']) keys.push('setup.status.' + s);
  for (const c of ['sms', 'email']) keys.push('setup.' + c + '.need', 'setup.' + c + '.no');
  for (const l of ['good', 'large', 'poor']) keys.push('brand.contrast.' + l);
  for (const k of keys) {
    assert.ok(m.fa[k], 'fa ' + k);
    assert.ok(m.en[k], 'en ' + k);
  }
  assert.ok(m.fa['brand.signIn'].includes('{name}') && m.en['brand.signIn'].includes('{name}'));
});

test('project settings sections are one registry with brand and setup first', () => {
  const settings = fs.readFileSync(path.join(ROOT, 'src', 'settings.js'), 'utf8');
  const ids = [...settings.matchAll(/^  \{ id: '([a-z-]+)', label: '([a-z.-]+)', icon: '([a-z]+)', acl: '([A-Z_]+)'/gm)].map((m) => m[1]);
  assert.deepEqual(ids, ['brand', 'setup', 'languages', 'datetime', 'countries', 'accounts', 'friends', 'otp', 'app-versions', 'categories', 'tags', 'payment-methods', 'sms', 'email', 'push', 'voice']);
  const pages = fs.readFileSync(path.join(ROOT, 'src', 'pages.js'), 'utf8');
  assert.doesNotMatch(pages, /SETTINGS_TABS|MOVED_PAGES/);
  assert.match(settings, /'settings\/sms': 'sms'/);
  assert.match(settings, /'base\/categories': 'categories'/);
  const setup = fs.readFileSync(path.join(ROOT, 'src', 'setup.js'), 'utf8');
  assert.match(setup, /var SETUP_ORDER = \['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review'\];/);
  assert.match(setup, /'\/setup\/claim'/);
});
