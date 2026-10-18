import assert from 'node:assert/strict';
import test from 'node:test';
import { JSDOM } from 'jsdom';
import { loadCatalog } from '../scripts/catalog.mjs';
import { loadRuntime } from '../scripts/audit.mjs';

const runtime = loadRuntime();
const index = runtime.compile(loadCatalog('en'), loadCatalog('fa-IR'), 'fa-IR');
const t = (text, el) => runtime.translate(index, text, el);
const LRI = '⁦';
const FSI = '⁨';
const PDI = '⁩';

test('exact messages translate and keep surrounding whitespace', () => {
  assert.equal(t('Players'), 'بازیکنان');
  assert.equal(t('  Sign in '), '  ورود ');
  assert.equal(t('Sign\n   in'), 'ورود');
});

test('unknown text is left alone', () => {
  assert.equal(t('player_1'), null);
  assert.equal(t('Some text that is not in the catalog'), null);
});

test('placeholders are interpolated and numbers localized', () => {
  assert.equal(t('1 to 7'), '۱ تا ۷');
  assert.equal(t('Players can have a max of 25 items.'), 'بازیکنان می‌توانند حداکثر ۲۵ عدد از این آیتم داشته باشند.');
  assert.equal(t('Signing in as admin'), `در حال ورود با حساب ${FSI}admin${PDI}`);
});

test('typed placeholders do not swallow arbitrary sentences', () => {
  // "{from} to {to}" only accepts numbers.
  assert.equal(t('Why lock-in matters to studios'), null);
  // "in {time}" only accepts known durations.
  assert.equal(t('in progress'), null);
  assert.equal(t('in 3 days'), '۳ روز دیگر');
});

test('plural source forms map to the Persian form', () => {
  assert.equal(t('1 minute'), '۱ دقیقه');
  assert.equal(t('12 minutes'), '۱۲ دقیقه');
  assert.equal(t('about 1 hour'), 'حدود ۱ ساعت');
  assert.equal(t('3 minutes ago'), '۳ دقیقه پیش');
});

test('nested placeholders translate recursively', () => {
  assert.equal(t('۲۹ دقیقه ago'), '۲۹ دقیقه پیش');
  assert.equal(t('Select or insert an energy ID'), 'یک شناسهٔ انرژی انتخاب یا وارد کنید');
});

test('context entries apply only inside matching elements', () => {
  const { document } = new JSDOM('<div class="x-toggle"><span id="a">read</span></div><p id="b">read</p>').window;
  assert.equal(t('read', document.getElementById('a')), 'خواندن');
  assert.equal(t('read', document.getElementById('b')), null);
  // Generic "Set" is the leaderboard operator; inside a label it is an item set.
  const { document: d2 } = new JSDOM('<label><span id="s">Set</span></label>').window;
  assert.equal(t('Set'), 'جایگزینی');
  assert.equal(t('Set', d2.getElementById('s')), 'مجموعه');
});

test('pure numbers become Persian digits with Persian separators', () => {
  assert.equal(runtime.toPersianDigits('1,234.5'), '۱٬۲۳۴٫۵');
  assert.equal(runtime.toPersianDigits('42%'), '۴۲٪');
  assert.equal(t('1,234,567'), '۱٬۲۳۴٬۵۶۷');
  assert.ok(runtime.isNumeric('3.5k'));
  assert.ok(!runtime.isNumeric('3.41.0'));
  assert.ok(!runtime.isNumeric('1234567890123456'));
});

test('fragments that start with punctuation attach to the previous element', () => {
  assert.equal(t(' if you are unable to access your account. '), '. ');
});

test('technical tokens inside Persian sentences stay in order', () => {
  const out = t("Set User ID to '00000000-0000-0000-0000-000000000000' or leave empty to make it a system owned storage object.");
  assert.ok(out.includes(`${LRI}00000000-0000-0000-0000-000000000000${PDI}`), out);
});

test('server messages shown in the console are translated', () => {
  assert.equal(t('Invalid credentials.'), 'اطلاعات ورود نامعتبر است.');
  assert.equal(t("Wallet rejected negative value at path 'coins'."), `کیف پول مقدار منفی در مسیر «${FSI}coins${PDI}» را نپذیرفت.`);
  assert.equal(t('player_1 wants to add you as a friend'), `${FSI}player_1${PDI} می‌خواهد شما را به‌عنوان دوست اضافه کند`);
});

test('document titles translate each breadcrumb segment', () => {
  assert.equal(runtime.translateTitle(index, 'Players | Nakama (127)'), 'بازیکنان | Nakama (127)');
  assert.equal(runtime.translateTitle(index, 'Settings - Users | Nakama'), 'تنظیمات - کاربران | Nakama');
  assert.equal(runtime.translateTitle(index, 'Nakama'), 'Nakama');
});

test('compile reports keys without translations', () => {
  const idx = runtime.compile({ ns: { a: 'A', b: 'B' } }, { ns: { a: 'آ' } }, 'fa-IR');
  assert.deepEqual([...idx.missing], ['ns.b']);
  assert.equal(runtime.translate(idx, 'B'), null);
});
