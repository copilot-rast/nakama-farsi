const L = require('./lib.cjs');
const WELCOME = { fa: 'به پنل مدیریت ستاره‌چین خوش آمدید.\nبرای ورود از حساب سازمانی خود استفاده کنید.', en: 'Welcome to the Starpicker admin panel.' };

(async () => {
  const b = await L.browser();
  const a = await L.context(b);
  await L.login(a.page);
  const saved = (await L.adminFetch('GET', '/v2/console/base/brand')).data;
  L.check(saved.custom === true && saved.name.fa, 'run walk.cjs first: the brand is set');
  const BRAND = { name: saved.name, shortName: saved.shortName, logoUrl: saved.logoUrl, logoDarkUrl: saved.logoDarkUrl, markUrl: saved.markUrl, accent: saved.accent, welcome: WELCOME };
  const put = await L.api(a.page, 'PUT', '/v2/console/base/brand', BRAND);
  L.check(put.status === 200, 'brand with welcome saved');
  for (const [loc, mode] of [['fa-IR', 'light'], ['en', 'dark']]) {
    const c = await L.context(b, { locale: loc, mode });
    await c.page.goto(L.URL + '/#/login');
    await c.page.waitForSelector('[data-brand-signin]', { timeout: 8000 });
    await c.page.waitForTimeout(800);
    const txt = await c.page.evaluate(() => document.querySelector('[data-brand-signin]').innerText);
    L.check(loc === 'en' ? /Starpicker/.test(txt) && /Welcome/.test(txt) : /ستاره‌چین/.test(txt) && /خوش آمدید/.test(txt), 'login shows brand + welcome ' + loc);
    await L.shot(c.page, `login-branded-welcome-${loc === 'en' ? 'en' : 'fa'}-${mode}`);
    await L.login(c.page);
    await c.page.waitForTimeout(1200);
    await L.shot(c.page, `sidebar-branded-${loc === 'en' ? 'en' : 'fa'}-${mode}`);
    if (mode === 'light') {
      await c.page.click('.x-drawer > div:first-child button');
      await c.page.waitForTimeout(800);
      await L.shot(c.page, 'sidebar-collapsed-fa-light');
      await c.page.click('.x-drawer > div:first-child button').catch(() => {});
    }

  }
  const m = await L.context(b, { mobile: true });
  await m.page.goto(L.URL + '/#/login');
  await m.page.waitForSelector('[data-brand-signin]');
  await L.shot(m.page, 'login-branded-fa-mobile');
  const back = await L.api(a.page, 'PUT', '/v2/console/base/brand', Object.assign({}, BRAND, { welcome: {} }));
  L.check(back.status === 200, 'welcome text removed again');
  await b.close();
})().catch((e) => { console.error(e); process.exit(1); });
