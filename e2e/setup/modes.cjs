const L = require('./lib.cjs');

const MODES = (process.argv[2] || 'fa-light,fa-dark,en-light,fa-mobile').split(',');
const STEPS = ['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review'];

(async () => {
  for (const mode of MODES) {
    const b = await L.browser();
    const [loc, look] = mode.split('-');
    const mobile = look === 'mobile';
    const { ctx, page } = await L.context(b, { locale: loc === 'en' ? 'en' : 'fa-IR', mode: look === 'dark' ? 'dark' : 'light', mobile });
    await page.goto(L.URL + '/#/login');
    await page.waitForSelector('input[name="username"]');
    await page.waitForTimeout(1200);
    await L.shot(page, `login-${mode}`);
    await L.login(page);
    await page.waitForTimeout(1500);
    L.check(!(await L.wizardOpen(page)), `${mode}: wizard does not auto-open again`);
    if (!mobile) await L.shot(page, `sidebar-${mode}`);
    await page.evaluate(() => { location.hash = '#/settings/setup'; });
    await page.waitForSelector('[data-setup-open]', { timeout: 10000 });
    await page.waitForTimeout(800);
    await L.shot(page, `settings-setup-${mode}`);
    await page.click('[data-setup-open]');
    await page.waitForSelector('[data-economy="setup"]');
    for (let i = 0; i < STEPS.length; i++) {
      const k = STEPS[i];
      await page.click(`[data-economy="setup"] .setup-step[data-step="${k}"]`);
      await L.waitStep(page, k);
      if (k === 'payment' || k === 'sms' || k === 'email') await page.waitForSelector('[data-economy="setup"] [data-need]', { timeout: 8000 }).catch(() => {});
      await page.waitForTimeout(500);
      await L.shot(page, `step${i + 1}-${k}-${mode}`);
    }
    if (mode === 'fa-light') {
      await page.keyboard.press('Escape');
      await page.waitForSelector('[data-economy="dialog"] [role="alertdialog"]', { timeout: 4000 });
      await L.shot(page, `esc-confirm-${mode}`);
      await page.keyboard.press('Escape');
      await page.waitForTimeout(300);
      L.check(await L.wizardOpen(page), 'Esc on the confirm keeps the wizard open');
      const ids = [];
      for (let i = 0; i < 40; i++) {
        await page.keyboard.press('Tab');
        ids.push(await page.evaluate(() => !!document.activeElement.closest('[data-economy="setup"]')));
      }
      L.check(ids.every(Boolean), 'Tab focus stays inside the wizard');
    }
    await ctx.close();
    await b.close();
  }
})().catch((e) => { console.error(e); process.exit(1); });
