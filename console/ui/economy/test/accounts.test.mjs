import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { ROOT } from '../scripts/build.mjs';

const src = fs.readFileSync(path.join(ROOT, 'src', 'accounts.js'), 'utf8');
const i18n = fs.readFileSync(path.join(ROOT, 'src', 'i18n.js'), 'utf8');

function pick(names) {
  return names.map((n) => {
    const fn = new RegExp(`^function ${n}\\([^)]*\\) \\{[\\s\\S]*?^\\}`, 'm').exec(src);
    if (fn) return fn[0];
    const v = new RegExp(`^var ${n} = .*;$`, 'm').exec(src);
    assert.ok(v, n);
    return v[0];
  }).join('\n');
}

const same = (a, b) => assert.deepEqual(JSON.parse(JSON.stringify(a)), b);
const ctx = { t: (k, v) => k + (v && v.n ? ':' + v.n : '') };
vm.runInNewContext(pick(['VERSION_RE', 'flagOf', 'versionParts', 'compareVersions', 'ruleProblems']) + ';this.flagOf = flagOf; this.compareVersions = compareVersions; this.ruleProblems = ruleProblems; this.VERSION_RE = VERSION_RE;', ctx);

test('versions compare like the server', () => {
  assert.equal(ctx.compareVersions('1.2.10', '1.2.9'), 1);
  assert.equal(ctx.compareVersions('1.2', '1.2.0'), 0);
  assert.equal(ctx.compareVersions('2.0.0-beta', '2.0.0'), -1);
  assert.ok(ctx.VERSION_RE.test('1.4.0') && ctx.VERSION_RE.test('v2.0.0+45') && !ctx.VERSION_RE.test('1..2') && !ctx.VERSION_RE.test('one'));
});

test('flags come from the country code', () => {
  assert.equal(ctx.flagOf('IR'), '🇮🇷');
  assert.equal(ctx.flagOf('de'), '🇩🇪');
  assert.equal(ctx.flagOf('USA'), '');
});

test('the rules preview matches the server rules', () => {
  const r = { usernameMin: 4, usernameMax: 8, usernameCharset: 'latin', passwordMin: 10, passwordLetter: true, passwordDigit: true };
  same(ctx.ruleProblems(r, 'abc', ''), ['code.USERNAME_TOO_SHORT:4']);
  same(ctx.ruleProblems(r, 'سارا_جان', ''), ['code.USERNAME_INVALID_CHARACTERS']);
  same(ctx.ruleProblems(r, '', 'abcdefghij'), ['code.PASSWORD_NEEDS_DIGIT']);
  same(ctx.ruleProblems(r, 'ali_99', 'abcdefghi1'), []);
});

test('every new section label and message exists in fa and en', () => {
  const m = new Function(i18n + ';return MESSAGES')();
  const keys = new Set();
  for (const mm of src.matchAll(/\bt\('([a-zA-Z0-9_.-]+)'/g)) keys.add(mm[1]);
  for (const mm of src.matchAll(/label: '([a-zA-Z0-9_.-]+)'/g)) keys.add(mm[1]);
  for (const mm of src.matchAll(/hint: '([a-zA-Z0-9_.-]+)'/g)) keys.add(mm[1]);
  for (const id of ['countries', 'accounts', 'otp', 'app-versions']) keys.add('tab.' + id).add('desc.' + id);
  for (const p of ['android', 'ios', 'web', 'windows', 'macos', 'linux']) keys.add('platform.' + p);
  for (const s of ['pending', 'verified', 'expired', 'failed', 'replaced', 'send_failed']) keys.add('otp.status.' + s);
  for (const k of keys) {
    if (k.endsWith('.')) continue;
    assert.ok(m.fa[k], 'fa ' + k);
    assert.ok(m.en[k], 'en ' + k);
  }
});
