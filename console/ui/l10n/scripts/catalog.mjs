// Catalog loading and validation shared by the build, audit and tests.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

export const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

export function readJSON(file) {
  return JSON.parse(fs.readFileSync(path.resolve(ROOT, file), 'utf8'));
}

export function loadConfig() {
  return readJSON('config.json');
}

export function loadCatalog(locale) {
  return readJSON(path.join('locales', `${locale}.json`));
}

// stripMeta removes "$schema"/"$meta" so only namespaces remain.
export function stripMeta(catalog) {
  const out = {};
  for (const [ns, entries] of Object.entries(catalog)) {
    if (!ns.startsWith('$')) out[ns] = entries;
  }
  return out;
}

export function messageOf(entry) {
  return typeof entry === 'string' ? entry : entry.message;
}

export function placeholders(text) {
  return [...new Set((text.match(/\{\w+\}/g) || []).sort())];
}

const ARABIC_YEH_KAF = /[يكى]/;
const BAD_ZWNJ = /(^|\s)‌|‌(\s|$)|‌‌/;

// validateCatalogs checks a translation catalog against the source catalog and
// returns {errors, missing, keys}. Missing translations are reported separately
// so the build can decide whether they are fatal.
export function validateCatalogs(source, target, { targetLocale = 'fa-IR' } = {}) {
  const errors = [];
  const missing = [];
  const seen = new Map();
  let keys = 0;
  const src = stripMeta(source);
  const dst = stripMeta(target);
  for (const ns of Object.keys(dst)) {
    if (!src[ns]) errors.push(`${targetLocale}: unknown namespace "${ns}"`);
  }
  for (const [ns, entries] of Object.entries(src)) {
    if (!/^[a-z][A-Za-z0-9]*$/.test(ns)) errors.push(`invalid namespace name "${ns}"`);
    const translations = dst[ns] || {};
    for (const key of Object.keys(translations)) {
      if (!(key in entries)) errors.push(`${targetLocale}: unknown key "${ns}.${key}"`);
    }
    for (const [key, entry] of Object.entries(entries)) {
      keys++;
      const id = `${ns}.${key}`;
      if (!/^[a-z][A-Za-z0-9]*$/.test(key)) errors.push(`invalid key "${id}"`);
      const message = messageOf(entry);
      if (typeof message !== 'string' || !message.trim()) {
        errors.push(`${id}: empty source message`);
        continue;
      }
      if (typeof entry === 'object' && (typeof entry.context !== 'string' || !entry.context.trim())) {
        errors.push(`${id}: context entries need a CSS selector "context"`);
      }
      const identity = `${message}\u0000${typeof entry === 'object' ? entry.context : ''}`;
      if (seen.has(identity)) errors.push(`${id}: duplicates source message of ${seen.get(identity)}`);
      seen.set(identity, id);

      const translation = translations[key];
      if (translation === undefined || translation === '') {
        missing.push(id);
        continue;
      }
      if (typeof translation !== 'string') {
        errors.push(`${id}: translation must be a string`);
        continue;
      }
      const srcForms = message.split(' | ');
      const dstForms = translation.split(' | ');
      if (dstForms.length > 2) errors.push(`${id}: at most two plural forms ("one | other") are supported`);
      const want = placeholders(srcForms.join(' '));
      const got = placeholders(dstForms.join(' '));
      if (want.join() !== got.join()) {
        errors.push(`${id}: placeholders differ (source ${want.join(',') || '-'}, translation ${got.join(',') || '-'})`);
      }
      if (targetLocale.startsWith('fa')) {
        if (ARABIC_YEH_KAF.test(translation)) errors.push(`${id}: uses Arabic Yeh/Kaf instead of Persian ی/ک`);
        if (BAD_ZWNJ.test(translation)) errors.push(`${id}: misplaced zero-width non-joiner`);
        if (/[ \t]{2,}/.test(translation)) errors.push(`${id}: repeated spaces`);
      }
    }
  }
  return { errors, missing, keys };
}
