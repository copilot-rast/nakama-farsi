const L = require('./lib.cjs');
const QA = { user: 'setupqa', pass: 'SetupQa-' + Math.random().toString(36).slice(2, 10) + '9!' };

(async () => {
  const b = await L.browser();
  const a = await L.context(b);
  await L.login(a.page);
  await a.page.waitForTimeout(1500);
  L.check(!(await L.wizardOpen(a.page)), 'admin again: no auto-open');
  const before = (await L.api(a.page, 'GET', '/v2/console/base/setup')).data;

  const add = await L.api(a.page, 'POST', '/v2/console/user', { username: QA.user, email: 'setupqa@example.com', acl: { ACCOUNT: { read: true }, SETTINGS: { read: true, write: true, delete: true }, HIRO_ECONOMY: { read: true, write: true }, NOTIFICATION: { read: true, write: true } } });
  console.log('add user', add.status);
  await a.page.waitForTimeout(1500);
  { const rs = await L.api(a.page, 'POST', '/v2/console/user/' + QA.user + '/reset/password'); console.log('reset', rs.status); if (rs.data && rs.data.code) add.data = { token: rs.data.code }; }
  await a.page.waitForTimeout(1200);
  if (!add.data.token) { const rs = await L.api(a.page, 'POST', '/v2/console/user/' + QA.user + '/reset/password'); add.data = { token: rs.data.code }; console.log('reset', rs.status); }
  const pc = await a.page.evaluate(async ([code, password]) => (await fetch('/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code, password }) })).status, [add.data.token, QA.pass]);
  console.log('password set', pc);

  const q = await L.context(b);
  await L.login(q.page, QA).catch(async (e) => { await L.shot(q.page, 'qa-login-fail'); console.log(await q.page.evaluate(() => document.body.innerText.slice(0, 400))); throw e; });
  await q.page.waitForTimeout(1500);
  L.check(!(await L.wizardOpen(q.page)), 'second console user: no auto-open');
  await q.page.evaluate(() => { location.hash = '#/settings/setup'; });
  await q.page.waitForSelector('[data-setup-open]');
  await q.page.click('[data-setup-open]');
  await q.page.waitForSelector('[data-economy="setup"] [data-setup-step]');
  L.check(true, 'second user reopens from #/settings/setup');
  await L.shot(q.page, 'second-user-reopen-fa-light');
  await q.page.click('[data-economy="setup"] [data-setup-later]');
  await q.page.waitForTimeout(600);
  const still = (await L.api(a.page, 'GET', '/v2/console/base/setup')).data;
  L.check(still.status === 'completed', 'later on a completed setup keeps it completed');

  await L.api(a.page, 'PUT', '/v2/console/base/setup', { status: 'pending' });
  const c1 = await L.context(b);
  const c2 = await L.context(b);
  const fail = (pg, n) => async (e) => { await L.shot(pg, 'conc-fail-' + n); console.log(n, pg.url(), (await pg.evaluate(() => document.body.innerText.slice(0, 200)))); throw e; };
  await Promise.all([L.login(c1.page).catch(fail(c1.page, 'admin')), L.login(c2.page, QA).catch(fail(c2.page, 'qa'))]);
  await Promise.all([c1.page.waitForTimeout(2500), c2.page.waitForTimeout(2500)]);
  const o1 = await L.wizardOpen(c1.page);
  const o2 = await L.wizardOpen(c2.page);
  const st = (await L.api(a.page, 'GET', '/v2/console/base/setup')).data;
  console.log('simultaneous logins: admin', o1, 'qa', o2, 'firstOpenedBy', st.firstOpenedBy);
  L.check(o1 !== o2, 'two simultaneous logins: exactly one gets the wizard');
  const winner = o1 ? c1.page : c2.page;
  await L.shot(winner, 'concurrency-winner');
  await winner.click('[data-economy="setup"] [data-setup-later]');
  await winner.waitForTimeout(500);
  await L.api(a.page, 'PUT', '/v2/console/base/setup', { steps: before.steps, status: 'completed' });

  const key = L.SERVER_KEY;
  const dev = await fetch(L.API_URL + '/v2/account/authenticate/device?create=true', { method: 'POST', headers: { Authorization: 'Basic ' + Buffer.from(key + ':').toString('base64'), 'Content-Type': 'application/json' }, body: JSON.stringify({ id: 'setup-qa-device-0001' }) }).then((r) => r.json());
  const rpc = async (payload) => fetch(L.API_URL + '/v2/rpc/base_brand?unwrap=true', { method: 'POST', headers: { Authorization: 'Bearer ' + dev.token, 'Content-Type': 'application/json' }, body: JSON.stringify(payload) }).then((r) => r.json());
  const fa = await rpc({ lang: 'fa' });
  const en = await rpc({ lang: 'en' });
  console.log('base_brand fa', JSON.stringify(fa));
  console.log('base_brand en', en.name, en.accent);
  L.check(fa.name === 'ستاره‌چین' && en.name === 'Starpicker' && fa.accent === '#059669' && fa.custom === true, 'base_brand RPC returns the brand to a device user');
  const anon = await fetch(L.API_URL + '/v2/rpc/base_brand?unwrap=true', { method: 'POST', body: '{}' });
  L.check(anon.status === 401, 'base_brand needs a session');

  await a.page.evaluate(() => { location.hash = '#/settings/brand'; });
  await a.page.waitForSelector('.brand-editor input');
  await a.page.waitForTimeout(600);
  await L.shot(a.page, 'settings-brand-fa-light');
  const reset = await L.api(a.page, 'PUT', '/v2/console/base/brand', {});
  L.check(reset.status === 200 && reset.data.custom === false, 'brand reset to empty');
  const fresh = await L.context(b);
  await fresh.page.goto(L.URL + '/#/login');
  await fresh.page.waitForSelector('input[name="username"]');
  await fresh.page.waitForTimeout(1200);
  const loginState = await fresh.page.evaluate(() => ({ title: document.title, ours: !!document.querySelector('[data-brand-logo],[data-brand-signin]'), h1: document.querySelector('h1').textContent }));
  console.log('login after reset', JSON.stringify(loginState));
  L.check(!loginState.ours && /Nakama/.test(loginState.h1), 'login page back to Nakama defaults');
  await L.shot(fresh.page, 'login-after-reset-fa-light');
  await L.login(fresh.page);
  await fresh.page.waitForTimeout(1500);
  const home = await fresh.page.evaluate(() => ({ title: document.title, icon: document.querySelector('link[rel~="icon"]').getAttribute('href'), p500: getComputedStyle(document.documentElement).getPropertyValue('--primary-color-500').trim(), style: !!document.getElementById('brand-accent'), native: [...document.querySelectorAll('.x-drawer a[href="#/"] img')].some((i) => i.offsetParent !== null) }));
  console.log('home after reset', JSON.stringify(home));
  L.check(/Nakama/.test(home.title) && home.icon === '/favicon.ico' && !home.style && home.native && /107\s+89\s+238/.test(home.p500), 'console back to Nakama defaults');
  await L.shot(fresh.page, 'dashboard-after-reset-fa-light');

  const del = await L.api(a.page, 'DELETE', '/v2/console/user?username=' + QA.user);
  console.log('qa user deleted', del.status);
  await b.close();
})().catch((e) => { console.error(e); process.exit(1); });
