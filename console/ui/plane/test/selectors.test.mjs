import assert from 'node:assert/strict';
import test from 'node:test';
import { bundle, loadSelectors, missingSelectors, nativeSources, unlistedClasses } from '../scripts/build.mjs';

const manifest = loadSelectors();

test('every manifest entry is described and has classes', () => {
  const ids = new Set();
  for (const e of manifest.selectors) {
    assert.ok(e.id && !ids.has(e.id), 'unique id ' + e.id);
    ids.add(e.id);
    assert.ok(e.selector && e.what && Array.isArray(e.classes) && e.classes.length, e.id);
    assert.ok(Array.isArray(e.routes) && Number.isInteger(e.min), e.id);
  }
});

test('the check fails when a native class disappears from the bundle', () => {
  const fixture = { css: '.rounded-md{}.shadow{}', js: 'av(`TableHead`) "sticky top-0"' };
  const m = { own: manifest.own, selectors: [{ id: 'card', classes: ['rounded-md', 'shadow', 'bg-white'] }, { id: 'table', classes: ['x-table-head', 'sticky'] }, { id: 'drawer', classes: ['x-drawer'] }] };
  assert.deepEqual(missingSelectors(fixture, m), ['card: .bg-white', 'drawer: .x-drawer']);
});

test('the check fails when plane css uses a native class that is not listed', () => {
  assert.deepEqual(unlistedClasses('html.pl .x-new-thing{} html.pl .pl-rail{} html.pl .eco-tag{}', manifest), ['x-new-thing']);
});

test('the real console bundle has every class plane depends on', () => {
  assert.deepEqual(missingSelectors(nativeSources(), manifest), []);
  assert.deepEqual(unlistedClasses(bundle().css, manifest), []);
});
