const { chromium } = require(process.env.PLAYWRIGHT_CORE || '/Users/mazyar/Projects/ai_chat/frontend/node_modules/playwright-core');
const fs = require('fs');
const path = require('path');
const URL = process.env.E2E_CONSOLE_URL || 'http://100.106.1.79:7451';
const API_URL = process.env.E2E_API_URL || 'http://100.106.1.79:7450';
const SHOTS = process.env.E2E_SHOTS || path.resolve(__dirname, '..', '..', '.local', 'shots-setup');
const ASSETS = path.join(__dirname, 'assets');
fs.mkdirSync(SHOTS, { recursive: true });
const ENV_FILE = process.env.E2E_ENV_FILE || process.env.HOME + '/.config/nakama-farsi/mac-dev.env';
const env = fs.readFileSync(ENV_FILE, 'utf8');
const SERVER_KEY = /SERVER_KEY=(.*)/.exec(env)[1];
const ADMIN = { user: /CONSOLE_USER=(.*)/.exec(env)[1], pass: /CONSOLE_PASSWORD=(.*)/.exec(env)[1] };

async function browser() {
  return chromium.launch({ executablePath: process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
}

async function context(b, opts = {}) {
  const ctx = await b.newContext({ viewport: opts.mobile ? { width: 390, height: 844 } : { width: 1440, height: 900 }, deviceScaleFactor: opts.mobile ? 2 : 1, isMobile: !!opts.mobile, hasTouch: !!opts.mobile });
  await ctx.addInitScript(([locale, mode]) => {
    try {
      if (!sessionStorage.getItem('x-init')) {
        if (locale) localStorage.setItem('nakama.console.locale', locale);
        if (mode) localStorage.setItem('color-mode', mode);
        sessionStorage.setItem('x-init', '1');
      }
    } catch (e) {}
  }, [opts.locale || 'fa-IR', opts.mode || 'light']);
  const page = await ctx.newPage();
  page.on('pageerror', (e) => console.log('PAGEERROR', e.message));
  page.on('console', (m) => { if (m.type() === 'error' && !/favicon|segment|heroiclabs/.test(m.text())) console.log('CONSOLE', m.text().slice(0, 200)); });
  return { ctx, page };
}

async function login(page, who = ADMIN) {
  await page.goto(URL + '/#/login');
  await page.waitForSelector('input[name="username"]');
  await page.waitForTimeout(1000);
  await page.fill('input[name="username"]', who.user);
  await page.fill('input[name="password"]', who.pass);
  await page.keyboard.press('Enter');
  await page.waitForSelector('.x-drawer a[href="#/players"]', { timeout: 15000, state: 'attached' });
}

async function shot(page, name, full) {
  await page.waitForTimeout(400);
  await page.screenshot({ path: path.join(SHOTS, name + '.png'), fullPage: !!full });
  console.log('shot', name);
}

async function token(page) {
  return page.evaluate(() => JSON.parse(localStorage.getItem('SESSION_TOKENS') || sessionStorage.getItem('SESSION_TOKENS') || '{}').accessToken);
}

async function api(page, method, p, body) {
  return page.evaluate(async ([method, p, body]) => {
    const tk = JSON.parse(localStorage.getItem('SESSION_TOKENS') || sessionStorage.getItem('SESSION_TOKENS') || '{}').accessToken;
    const r = await fetch(p, { method, headers: { Authorization: 'Bearer ' + tk, 'Content-Type': 'application/json' }, body: body === undefined ? undefined : JSON.stringify(body) });
    const text = await r.text();
    let data = null;
    try { data = JSON.parse(text); } catch (e) { data = text; }
    return { status: r.status, data };
  }, [method, p, body]);
}

async function wizardOpen(page) {
  return page.evaluate(() => !!document.querySelector('[data-economy="setup"]'));
}

async function stepTitle(page) {
  return page.evaluate(() => { const e = document.querySelector('[data-economy="setup"] #setup-step-title'); return e ? e.textContent : null; });
}

async function waitStep(page, k) {
  await page.waitForSelector(`[data-economy="setup"] [data-setup-step="${k}"]`, { timeout: 10000 });
  await page.waitForTimeout(900);
}

async function next(page) {
  await page.click('[data-economy="setup"] [data-setup-next]');
}

async function adminFetch(method, p, body) {
  const r = await fetch(URL + p, { method, headers: { Authorization: 'Basic ' + Buffer.from(ADMIN.user + ':' + ADMIN.pass).toString('base64'), 'Content-Type': 'application/json' }, body: body === undefined ? undefined : JSON.stringify(body) });
  const text = await r.text();
  let data = null;
  try { data = JSON.parse(text); } catch (e) { data = text; }
  return { status: r.status, data };
}

function check(cond, msg) {
  console.log((cond ? 'PASS ' : 'FAIL ') + msg);
  if (!cond) process.exitCode = 1;
}

module.exports = { adminFetch, API_URL, SERVER_KEY, browser, context, login, shot, token, api, wizardOpen, stepTitle, waitStep, next, check, URL, ADMIN, ASSETS, SHOTS };
