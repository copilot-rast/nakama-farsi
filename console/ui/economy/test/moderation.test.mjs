import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { ROOT, SOURCES } from '../scripts/build.mjs';

const src = fs.readFileSync(path.join(ROOT, 'src', 'moderation.js'), 'utf8');
const i18n = fs.readFileSync(path.join(ROOT, 'src', 'i18n.js'), 'utf8');
const m = new Function(i18n + ';return MESSAGES')();

test('the moderation page is bundled after the pages it builds on', () => {
  assert.ok(SOURCES.indexOf('moderation.js') > SOURCES.indexOf('publication.js'));
  assert.ok(SOURCES.indexOf('moderation.js') < SOURCES.indexOf('shell.js'));
});

test('every moderation label, message and dynamic key exists in fa and en', () => {
  const keys = new Set();
  for (const mm of src.matchAll(/\bt\('([a-zA-Z0-9_.-]+)'/g)) keys.add(mm[1]);
  for (const mm of src.matchAll(/(?:label|hint): '([a-zA-Z0-9_.-]+)'/g)) keys.add(mm[1]);
  for (const kind of ['suspend', 'mute', 'warn']) {
    for (const p of ['rs.dialog.', 'rs.do.', 'rs.done.', 'rs.kind.', 'rs.reason.hint.', 'rs.reason.default.', 'rp.action.', 'rp.submit.', 'rp.done.']) keys.add(p + kind);
    keys.add('rs.dialog.' + kind + '.text');
  }
  for (const a of ['dismiss', 'delete']) for (const p of ['rp.action.', 'rp.submit.', 'rp.done.']) keys.add(p + a);
  for (const a of ['dismiss', 'delete', 'warn', 'mute', 'suspend']) keys.add('rp.action.' + a + '.hint');
  for (const s of ['open', 'dismissed', 'resolved']) keys.add('rp.status.' + s);
  for (const s of ['active', 'expired', 'lifted']) keys.add('rs.status.' + s);
  for (const s of ['suspended', 'muted', 'banned', 'tagged']) keys.add('pl.status.' + s);
  for (const a of ['mask', 'block', 'flag']) keys.add('cw.action.' + a);
  for (const a of ['word', 'contains']) keys.add('cw.match.' + a);
  for (const p of ['players', 'reports', 'restrictions', 'chat-words']) keys.add('nav.' + p).add('desc.' + p).add('empty.hint.' + p);
  keys.add('section.mod').add('section.mod.short').add('tab.report-titles').add('desc.report-titles').add('rp.openN');
  for (const k of keys) {
    if (k.endsWith('.')) continue;
    assert.ok(m.fa[k], 'fa ' + k);
    assert.ok(m.en[k], 'en ' + k);
  }
});

test('Persian moderation strings use ZWNJ and never the Arabic ي or ك', () => {
  const fa = Object.entries(m.fa).filter(([k]) => /^(rs|rp|pl|pc|cw|dur|tag)\.|^section\.mod|^nav\.(players|reports|restrictions|chat-words)/.test(k));
  assert.ok(fa.length > 100);
  for (const [k, v] of fa) {
    if (k !== 'cw.try.hint' && k !== 'f.cw.word.hint') assert.ok(!/[يك]/.test(v), k + ' has Arabic letters');
    assert.ok(!/(^|[\s«])ن?می /.test(v), k + ' needs a ZWNJ after می');
  }
});

test('durations read as hours and days in the console language', () => {
  const ctx = { t: (k, v) => k + ':' + v.n, localDigits: (s) => String(s) };
  vm.runInNewContext(/^function durationLabel\([^)]*\) \{[\s\S]*?^\}/m.exec(src)[0] + ';this.durationLabel = durationLabel;', ctx);
  assert.equal(ctx.durationLabel(60), 'dur.hours:1');
  assert.equal(ctx.durationLabel(4320), 'dur.days:3');
});
