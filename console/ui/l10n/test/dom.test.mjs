// Runs the built runtime (dist/static/l10n-*.js) in a simulated browser.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { JSDOM } from 'jsdom';
import { ROOT } from '../scripts/catalog.mjs';

const DIST = path.resolve(ROOT, '..', 'dist');
const html = fs.readFileSync(path.join(DIST, 'index.html'), 'utf8');
const runtimeSrc = /<script src="(\/static\/l10n-[a-f0-9]+\.js)"><\/script>/.exec(html)[1];
const runtimeCode = fs.readFileSync(path.join(DIST, runtimeSrc), 'utf8');

function boot({ url = 'http://console.test/', locale, body = '', beforeRuntime } = {}) {
  const dom = new JSDOM(`<!DOCTYPE html><html lang="en"><head><title>Nakama</title></head><body>${body}</body></html>`, {
    url,
    runScripts: 'outside-only',
    pretendToBeVisual: true,
  });
  if (locale) dom.window.localStorage.setItem('nakama.console.locale', locale);
  if (beforeRuntime) beforeRuntime(dom.window);
  dom.window.eval(runtimeCode);
  return dom.window;
}

const tick = (win) => new Promise((resolve) => win.setTimeout(resolve, 20));

test('index.html loads the runtime before the console bundle', () => {
  assert.match(html, /<html lang="fa-IR" dir="rtl">/);
  const runtimeAt = html.indexOf(runtimeSrc);
  const bundleAt = html.search(/<script type="module"[^>]*src="\/static\/index-/);
  assert.ok(runtimeAt > 0 && runtimeAt < bundleAt);
  assert.match(html, /href="\/static\/index-[^"]+\.logical-[a-f0-9]+\.css" data-l10n-source="\/static\/index-[^"]+\.css"/);
});

test('Persian is the default: root is fa-IR and right-to-left', () => {
  const win = boot();
  assert.equal(win.document.documentElement.getAttribute('lang'), 'fa-IR');
  assert.equal(win.document.documentElement.getAttribute('dir'), 'rtl');
  assert.equal(win.NakamaL10n.locale, 'fa-IR');
});

test('the console cannot reset the root language back to English', async () => {
  const win = boot();
  win.document.documentElement.setAttribute('lang', 'en');
  await tick(win);
  assert.equal(win.document.documentElement.getAttribute('lang'), 'fa-IR');
});

test('rendered text, attributes and later updates are translated', async () => {
  const win = boot();
  const { document } = win;
  document.body.innerHTML = '<div id="app"><h1> Players </h1><input placeholder="Enter your username"><button title="Refresh">Refresh</button><span id="n">5 minutes ago</span></div>';
  await tick(win);
  assert.equal(document.querySelector('h1').textContent, ' بازیکنان ');
  assert.equal(document.querySelector('input').getAttribute('placeholder'), 'نام کاربری خود را وارد کنید');
  assert.equal(document.querySelector('button').getAttribute('title'), 'بازخوانی');
  assert.equal(document.getElementById('n').textContent, '۵ دقیقه پیش');
  // A framework updating the same text node must be re-translated.
  document.getElementById('n').firstChild.nodeValue = '6 minutes ago';
  await tick(win);
  assert.equal(document.getElementById('n').textContent, '۶ دقیقه پیش');
});

test('code, editors and user input are never rewritten', async () => {
  const win = boot();
  const { document } = win;
  document.body.innerHTML = '<pre>Players</pre><code>Save</code><div class="monaco-editor"><span>Delete</span></div><textarea>Cancel</textarea><span translate="no">Edit</span>';
  await tick(win);
  assert.equal(document.querySelector('pre').textContent, 'Players');
  assert.equal(document.querySelector('code').textContent, 'Save');
  assert.equal(document.querySelector('.monaco-editor span').textContent, 'Delete');
  assert.equal(document.querySelector('textarea').value, 'Cancel');
  assert.equal(document.querySelector('[translate=no]').textContent, 'Edit');
});

test('identifiers keep Latin digits; untranslated English prose is marked LTR', async () => {
  const win = boot();
  const { document } = win;
  document.body.innerHTML = '<table><tr><td class="font-mono">1500</td></tr></table><p id="prose">A custom backend is a hiring spec.</p><span id="cron">0 0 * * 1</span><span id="uuid">807c89c8-a27f-466e</span>';
  await tick(win);
  assert.equal(document.querySelector('.font-mono').textContent, '1500');
  assert.ok(document.getElementById('prose').hasAttribute('data-l10n-ltr'));
  assert.equal(document.getElementById('cron').textContent, '⁦0 0 * * 1⁩');
  assert.equal(document.getElementById('uuid').textContent, '807c89c8-a27f-466e');
});

test('directional icons are marked for mirroring, others are not', async () => {
  const win = boot();
  const { document } = win;
  document.body.innerHTML =
    '<svg id="next" viewBox="0 0 24 24"><polyline points="9 18 15 12 9 6"></polyline></svg>' +
    '<svg id="search" viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"></circle></svg>';
  await tick(win);
  assert.equal(document.getElementById('next').getAttribute('data-l10n-icon'), 'mirror');
  assert.equal(document.getElementById('search').getAttribute('data-l10n-icon'), '');
});

test('document title is localized', async () => {
  const win = boot();
  win.document.title = 'Settings - Users | Nakama (localhost)';
  await tick(win);
  assert.equal(win.document.title, 'تنظیمات - کاربران | Nakama (localhost)');
});

test('default-locale dates use the Solar Hijri calendar; explicit locales are untouched', () => {
  const win = boot();
  const nowruz = new win.Date(Date.UTC(2024, 2, 20, 12));
  const fa = new win.Intl.DateTimeFormat(undefined, { dateStyle: 'long', timeZone: 'UTC' }).format(nowruz);
  assert.equal(fa, '۱ فروردین ۱۴۰۳');
  assert.equal(new win.Intl.DateTimeFormat(undefined, { timeZone: 'UTC' }).resolvedOptions().calendar, 'persian');
  // Leap year boundary: 1403 is a leap year in the Solar Hijri calendar (30 Esfand exists).
  assert.equal(new win.Intl.DateTimeFormat(undefined, { dateStyle: 'long', timeZone: 'UTC' }).format(new win.Date(Date.UTC(2025, 2, 20))), '۳۰ اسفند ۱۴۰۳');
  assert.equal(new win.Intl.DateTimeFormat('en-US', { timeZone: 'UTC' }).format(nowruz), '3/20/2024');
  assert.equal(nowruz.toLocaleString('en-US', { timeZone: 'UTC' }), '3/20/2024, 12:00:00 PM');
  assert.match(nowruz.toLocaleDateString(undefined, { timeZone: 'UTC' }), /۱۴۰۳/);
  assert.ok(new win.Intl.DateTimeFormat() instanceof win.Intl.DateTimeFormat);
});

test('number formatting: Persian digits, parse-safe formatToParts, currencies kept', () => {
  const win = boot();
  const nf = new win.Intl.NumberFormat();
  assert.equal(nf.format(1234567.5), '۱٬۲۳۴٬۵۶۷٫۵');
  const digits = nf.formatToParts(1234).filter((p) => p.type === 'integer').map((p) => p.value).join('');
  assert.equal(digits, '1234');
  assert.match(new win.Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(12.5), /\$۱۲٫۵۰/);
  assert.equal(new win.Intl.NumberFormat('en', { minimumFractionDigits: 1 }).format(2), '2.0');
});

test('relative times stay compatible with the console stripping " ago"', () => {
  const win = boot();
  const rtf = new win.Intl.RelativeTimeFormat('en', { numeric: 'always' });
  assert.equal(rtf.format(-5, 'minute'), '۵ دقیقه ago');
  assert.equal(rtf.format(-5, 'minute').replace(' ago', ''), '۵ دقیقه');
  assert.equal(rtf.format(3, 'day'), '۳ روز دیگر');
});

test('English can be selected and persists; nothing is translated or patched', async () => {
  const win = boot({ url: 'http://console.test/players?lang=en' });
  assert.equal(win.document.documentElement.getAttribute('dir'), 'ltr');
  assert.equal(win.document.documentElement.getAttribute('lang'), 'en');
  assert.equal(win.localStorage.getItem('nakama.console.locale'), 'en');
  win.document.body.innerHTML = '<h1>Players</h1>';
  await tick(win);
  assert.equal(win.document.querySelector('h1').textContent, 'Players');
  assert.equal(new win.Intl.NumberFormat().format(1234), '1,234');
});

test('a stored preference is honoured and unknown locales fall back to Persian', () => {
  assert.equal(boot({ locale: 'en' }).NakamaL10n.locale, 'en');
  assert.equal(boot({ locale: 'de' }).NakamaL10n.locale, 'fa-IR');
  assert.equal(boot({ url: 'http://console.test/?lang=fa' }).NakamaL10n.locale, 'fa-IR');
});

test('the language switcher toggles to the other locale', async () => {
  const win = boot();
  win.document.dispatchEvent(new win.Event('DOMContentLoaded'));
  await tick(win);
  const button = win.document.querySelector('.l10n-switcher');
  assert.ok(button);
  assert.equal(button.textContent, 'EN');
  assert.equal(button.getAttribute('lang'), 'en');
  assert.equal(button.getAttribute('aria-label'), 'Switch language to English');
});

test('the password-reset email opened via mailto: is localized', () => {
  const forwarded = [];
  const win = boot({ beforeRuntime: (w) => { w.open = (url) => { forwarded.push(url); return null; }; } });
  const body = 'Your Nakama Console password has been reset by a team member. Please click on the following link to set a new password: \n\n https://console.test/#/reset?token=t1 \n\n Thanks, \n Heroic Labs';
  win.open(`mailto:a@b.c?subject=Your Nakama Console password has been reset.&body=${encodeURIComponent(body)}`);
  assert.equal(forwarded.length, 1);
  const url = forwarded[0];
  assert.ok(url.startsWith('mailto:a@b.c?'));
  const params = new URLSearchParams(url.slice(url.indexOf('?') + 1));
  assert.equal(params.get('subject'), 'رمز عبور کنسول Nakama شما بازنشانی شد.');
  assert.match(params.get('body'), /^رمز عبور کنسول Nakama شما توسط یکی از اعضای تیم بازنشانی شد\./);
  assert.match(params.get('body'), /\n\nhttps:\/\/console\.test\/#\/reset\?token=t1\n\n/);
  assert.doesNotMatch(params.get('body'), /[\u2066-\u2069]/);
  // Other URLs pass through unchanged.
  win.open('https://heroiclabs.com/docs');
  assert.equal(forwarded[1], 'https://heroiclabs.com/docs');
});
