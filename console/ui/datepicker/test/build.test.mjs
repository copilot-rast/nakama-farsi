import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { CSS_SOURCES, END, JS_SOURCES, ROOT, START, bundle, injectIndex, restoreIndex } from '../scripts/build.mjs';
import * as economy from '../../economy/scripts/build.mjs';
import * as plane from '../../plane/scripts/build.mjs';

const vendor = '<!DOCTYPE html>\n<html lang="en">\n  <head>\n    <meta charset="UTF-8">\n    <script>\n      window.CONSOLE_CONFIG = {\n        nt: "{{nt}}",\n      }\n    </script>\n    <script type="module" crossorigin src="/static/index-A.js"></script>\n    <link rel="stylesheet" crossorigin href="/static/index-B.css">\n  </head>\n</html>\n';

test('the block goes right after the console config, before every other script, and comes out cleanly', () => {
  const once = injectIndex(vendor, '/static/datepicker-1.css', '/static/datepicker-1.js');
  assert.equal(once.split(START).length, 2);
  assert.ok(once.indexOf(START) > once.indexOf('CONSOLE_CONFIG'));
  assert.ok(once.indexOf(END) < once.indexOf('type="module"'));
  assert.ok(!/<script defer src="\/static\/datepicker/.test(once), 'the picker loads synchronously');
  assert.equal(injectIndex(once, '/static/datepicker-1.css', '/static/datepicker-1.js'), once);
  assert.equal(restoreIndex(once), vendor);
});

test('economy and plane builds keep the picker first in any order', () => {
  const steps = {
    dp: (h) => injectIndex(h, '/static/datepicker-1.css', '/static/datepicker-1.js'),
    economy: (h) => economy.injectIndex(h, '/static/economy-1.css', '/static/economy-1.js'),
    plane: (h) => plane.injectIndex(h, '/static/plane-1.css', '/static/plane-1.js'),
  };
  const orders = [['dp', 'economy', 'plane'], ['economy', 'dp', 'plane'], ['plane', 'economy', 'dp'], ['economy', 'plane', 'dp']];
  const results = orders.map((o) => o.reduce((h, s) => steps[s](h), vendor));
  for (const r of results) assert.equal(r, results[0]);
  const html = results[0];
  assert.ok(html.indexOf(END) < html.indexOf('economy-1.js'));
  assert.ok(html.indexOf(END) < html.indexOf('plane-1.js'));
});

test('the real index has the picker before the other bundles', () => {
  const html = fs.readFileSync(path.resolve(ROOT, '..', 'dist', 'index.html'), 'utf8');
  const s = html.indexOf(START);
  assert.ok(s !== -1, 'dist/index.html has the datepicker block');
  assert.ok(s < html.indexOf('<!-- nakama-l10n:start -->'));
  assert.ok(s < html.indexOf('type="module"'));
  const firstOther = html.search(/<!-- nakama-(?!datepicker)[a-z0-9]+:start -->/);
  assert.ok(firstOther === -1 || s < firstOther);
});

test('the bundle parses and no source has a comment', () => {
  const { js } = bundle();
  assert.doesNotThrow(() => new vm.Script(js));
  for (const f of [...JS_SOURCES, ...CSS_SOURCES]) {
    const src = fs.readFileSync(path.join(ROOT, 'src', f), 'utf8');
    assert.ok(!/^\s*\/\/|\/\*/m.test(src.replace(/'[^'\n]*'/g, "''").replace(/"[^"\n]*"/g, '""')), `${f} has a comment`);
  }
});
