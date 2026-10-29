import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { END, ROOT, START, bundle, injectIndex, restoreIndex } from '../scripts/build.mjs';

const L10N_BUILD = path.resolve(ROOT, '..', 'l10n', 'scripts', 'build.mjs');

const vendor = '<!DOCTYPE html>\n<html lang="en">\n  <head>\n    <script type="module" crossorigin src="/static/index-A.js"></script>\n    <link rel="stylesheet" crossorigin href="/static/index-B.css">\n  </head>\n</html>\n';
const l10nBlock = '    <!-- nakama-l10n:start -->\n    <script src="/static/l10n-1.js"></script>\n    <!-- nakama-l10n:end -->\n';

function messages() {
  const sandbox = {};
  vm.runInNewContext(fs.readFileSync(path.join(ROOT, 'src', 'i18n.js'), 'utf8') + ';this.MESSAGES = MESSAGES;', sandbox);
  return sandbox.MESSAGES;
}

test('the economy block is injected once and removed cleanly', () => {
  const once = injectIndex(vendor, '/static/economy-1.css', '/static/economy-1.js');
  assert.equal(once.split(START).length, 2);
  assert.ok(once.indexOf(END) < once.indexOf('<script type="module"'));
  assert.equal(injectIndex(once, '/static/economy-1.css', '/static/economy-1.js'), once);
  assert.equal(restoreIndex(once), vendor);
  assert.equal(restoreIndex(vendor), vendor);
});

test('the economy block sits before the l10n block so both builds agree in any order', async () => {
  const { restoreIndex: restoreL10n } = await import(L10N_BUILD);
  const withL10n = vendor.replace('    <script type="module"', l10nBlock + '    <script type="module"');
  const economyLast = injectIndex(withL10n, '/static/economy-1.css', '/static/economy-1.js');
  assert.ok(economyLast.indexOf(END) < economyLast.indexOf('<!-- nakama-l10n:start -->'));
  const l10nRebuilt = restoreL10n(economyLast).replace('    <script type="module"', l10nBlock + '    <script type="module"');
  assert.equal(l10nRebuilt.replace('<html lang="en">', '<html lang="en">'), economyLast);
  const economyFirst = injectIndex(vendor, '/static/economy-1.css', '/static/economy-1.js');
  const thenL10n = restoreL10n(economyFirst).replace('    <script type="module"', l10nBlock + '    <script type="module"');
  assert.equal(thenL10n, economyLast);
});

test('the bundle parses and contains no comments', () => {
  const { js, css } = bundle();
  assert.doesNotThrow(() => new vm.Script(js));
  for (const f of fs.readdirSync(path.join(ROOT, 'src'))) {
    const src = fs.readFileSync(path.join(ROOT, 'src', f), 'utf8');
    assert.ok(!/^\s*\/\/|\/\*/m.test(src.replace(/'[^'\n]*'/g, "''").replace(/"[^"\n]*"/g, '""')), `${f} has a comment`);
  }
  assert.ok(css.length > 0);
});

test('fa and en catalogs have the same keys', () => {
  const m = messages();
  assert.deepEqual(Object.keys(m.fa).sort(), Object.keys(m.en).sort());
});

test('Persian copy uses ZWNJ, Persian punctuation and no em dashes', () => {
  const m = messages();
  for (const [k, v] of Object.entries(m.fa)) {
    assert.ok(!/[—]/.test(v), `${k}: em dash`);
    assert.ok(!/(^|[\s(])(ن?می) [؀-ۿ]/.test(v), `${k}: "می" needs a ZWNJ: ${v}`);
    assert.ok(!/[؀-ۿ] (ها|های|هایی|تر|ترین)([\s.،؛:)]|$)/.test(v), `${k}: suffix needs a ZWNJ: ${v}`);
    assert.ok(!/[؀-ۿ],/.test(v), `${k}: use the Persian comma: ${v}`);
    assert.ok(!/[؀-ۿ]\?/.test(v), `${k}: use the Persian question mark: ${v}`);
    assert.ok(!/[0-9]/.test(v.replace(/\{\w+\}/g, '').replace(/[A-Za-z0-9_.\-()]+/g, '')), `${k}: Latin digit in Persian text: ${v}`);
  }
  for (const [k, v] of Object.entries(m.en)) assert.ok(!/[—]/.test(v), `${k}: em dash`);
});

test('committed dist output matches the sources', () => {
  execFileSync(process.execPath, [path.join(ROOT, 'scripts', 'build.mjs'), '--check'], { stdio: 'pipe' });
});

test('every literal message key exists and the removed economy taxonomy endpoints are gone', () => {
  const m = messages();
  const { js } = bundle();
  const keys = new Set([...js.matchAll(/\bt\('([a-zA-Z0-9_.-]*[a-zA-Z0-9_-])'\s*[,)]/g)].map((x) => x[1]));
  const missing = [...keys].filter((k) => m.en[k] === undefined);
  assert.deepEqual(missing, []);
  assert.ok(!/['"]\/categories['"]\s*\)|kind: '(product|prize|faq|news)'/.test(js));
  for (const code of ['DEFAULT_LANGUAGE_REQUIRED', 'INVALID_LANGUAGE', 'TEXT_TOO_LONG', 'TAG_KIND_MISMATCH', 'CATEGORY_KIND_MISMATCH']) assert.ok(m.fa['code.' + code] && m.en['code.' + code], code);
  assert.equal(m.fa['code.FA_REQUIRED'], undefined);
});
