import assert from 'node:assert/strict';
import test from 'node:test';
import { loadCatalog, loadConfig, stripMeta, validateCatalogs } from '../scripts/catalog.mjs';

const config = loadConfig();
const source = loadCatalog(config.sourceLocale);
const persian = loadCatalog('fa-IR');

test('fa-IR is the default locale and is right-to-left', () => {
  assert.equal(config.defaultLocale, 'fa-IR');
  assert.equal(config.locales['fa-IR'].dir, 'rtl');
  assert.equal(config.locales['fa-IR'].lang, 'fa-IR');
  assert.equal(config.locales.en.dir, 'ltr');
});

test('every source message has a valid Persian translation', () => {
  const { errors, missing, keys } = validateCatalogs(source, persian, { targetLocale: 'fa-IR' });
  assert.deepEqual(errors, []);
  assert.deepEqual(missing, []);
  assert.ok(keys > 1500, `expected a full catalog, got ${keys} keys`);
});

test('catalogs share the same namespaces and keys', () => {
  const src = stripMeta(source);
  const dst = stripMeta(persian);
  assert.deepEqual(Object.keys(dst), Object.keys(src));
  for (const ns of Object.keys(src)) assert.deepEqual(Object.keys(dst[ns]), Object.keys(src[ns]), ns);
});

test('catalog metadata describes each locale', () => {
  assert.equal(source.$meta.locale, 'en');
  assert.equal(persian.$meta.locale, 'fa-IR');
  assert.equal(persian.$meta.dir, 'rtl');
});

test('validator reports placeholder, orthography and key problems', () => {
  const src = { ns: { greet: 'Hello {name}', count: '1 item | {count} items', ok: 'Fine' } };
  const bad = {
    ns: {
      greet: 'سلام', // placeholder dropped
      count: 'يك {count} مورد', // Arabic Yeh/Kaf
      extra: 'اضافه',
    },
  };
  const { errors, missing } = validateCatalogs(src, bad, { targetLocale: 'fa-IR' });
  assert.ok(errors.some((e) => e.includes('ns.greet') && e.includes('placeholders differ')));
  assert.ok(errors.some((e) => e.includes('ns.count') && e.includes('Arabic Yeh/Kaf')));
  assert.ok(errors.some((e) => e.includes('unknown key "ns.extra"')));
  assert.deepEqual(missing, ['ns.ok']);
});

test('validator rejects misplaced zero-width non-joiners', () => {
  const { errors } = validateCatalogs({ ns: { a: 'A' } }, { ns: { a: 'می‌ ‌شود' } });
  assert.ok(errors.some((e) => e.includes('zero-width non-joiner')));
});

test('Persian translations use half-spaces in common compound words', () => {
  const all = JSON.stringify(stripMeta(persian));
  for (const word of ['می‌شود', 'ثبت‌نام', 'به‌روزرسانی', 'پیش‌فرض']) assert.ok(all.includes(word), word);
  for (const wrong of ['ميشود', 'می شود', 'بروزرسانی']) assert.ok(!all.includes(wrong), wrong);
});
