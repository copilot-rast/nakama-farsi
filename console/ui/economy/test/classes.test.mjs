import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { ROOT, SOURCES } from '../scripts/build.mjs';

const STATIC = path.resolve(ROOT, '..', 'dist', 'static');
const OWN = /^(eco|ps|setup|brand|economy|l10n|x)-|^(x|eco)$/;

function selectors(css) {
  const out = new Set();
  for (const m of css.matchAll(/\.((?:\\.|[A-Za-z0-9_-])+)/g)) out.add(m[1].replace(/\\(.)/g, '$1'));
  return out;
}

function classTokens(js) {
  const out = new Set();
  const lits = [];
  for (const m of js.matchAll(/class: '([^']*)'/g)) lits.push(m[1]);
  for (const m of js.matchAll(/classList\.(?:add|remove|toggle)\('([^']+)'/g)) lits.push(m[1]);
  for (const m of js.matchAll(/^var (?:NAV_[A-Z]+|BTN|INPUT(?:_SM|_MD)?|SELECT_[A-Z]+) = (?:INPUT \+ )?'([^']*)';/gm)) lits.push(m[1]);
  for (const m of js.matchAll(/class: '[^']*' \+ \(?[^'\n]*\? '([^']*)' : '([^']*)'/g)) lits.push(m[1], m[2]);
  for (const s of lits) for (const tok of s.split(/\s+/)) if (tok) out.add(tok);
  return out;
}

test('every utility class the overlay uses exists in the native bundle or economy.css', () => {
  const natives = fs.readdirSync(STATIC).filter((f) => /^index-[^.]+\.css$/.test(f));
  assert.ok(natives.length, 'native console css not found');
  const known = new Set();
  for (const f of natives) for (const s of selectors(fs.readFileSync(path.join(STATIC, f), 'utf8'))) known.add(s);
  for (const s of selectors(fs.readFileSync(path.join(ROOT, 'src', 'economy.css'), 'utf8'))) known.add(s);
  const missing = [];
  for (const f of SOURCES) {
    for (const tok of classTokens(fs.readFileSync(path.join(ROOT, 'src', f), 'utf8'))) {
      if (OWN.test(tok) || /^(primary|secondary|success|warning|error)$/.test(tok) || known.has(tok)) continue;
      missing.push(f + ': ' + tok);
    }
  }
  assert.deepEqual(missing, []);
});
