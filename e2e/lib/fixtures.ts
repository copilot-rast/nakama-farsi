import { test as base, chromium, expect, type Browser, type BrowserContext, type Page } from '@playwright/test';
import fs from 'node:fs';
import { AUTH_FILE, CHROME, CONSOLE_URL } from './env';

export type PageOpts = { locale?: string; mode?: 'light' | 'dark'; mobile?: boolean; signedIn?: boolean; width?: number; height?: number };

async function launch(): Promise<Browser> {
  return chromium.launch({ executablePath: CHROME, headless: true });
}

async function makePage(browser: Browser, opts: PageOpts): Promise<{ ctx: BrowserContext; page: Page }> {
  const mobile = !!opts.mobile;
  const ctx = await browser.newContext({
    viewport: { width: opts.width || (mobile ? 390 : 1440), height: opts.height || (mobile ? 844 : 900) },
    deviceScaleFactor: mobile ? 2 : 1,
    isMobile: mobile,
    hasTouch: mobile,
    locale: 'fa-IR',
    baseURL: CONSOLE_URL,
  });
  ctx.setDefaultTimeout(15_000);
  ctx.setDefaultNavigationTimeout(30_000);
  const saved: Record<string, string> = {};
  if (opts.signedIn) {
    const state = JSON.parse(fs.readFileSync(AUTH_FILE, 'utf8'));
    for (const o of state.origins || []) if (o.origin === new URL(CONSOLE_URL).origin) for (const kv of o.localStorage) saved[kv.name] = kv.value;
  }
  await ctx.addInitScript(([locale, mode, origin, items]) => {
    try {
      if (!sessionStorage.getItem('e2e-init')) {
        if (location.origin === origin) for (const k of Object.keys(items)) localStorage.setItem(k, items[k]);
        if (locale) localStorage.setItem('nakama.console.locale', locale);
        if (mode) localStorage.setItem('color-mode', mode);
        sessionStorage.setItem('e2e-init', '1');
      }
    } catch (e) {}
  }, [opts.locale || 'fa-IR', opts.mode || 'light', new URL(CONSOLE_URL).origin, saved] as [string, string, string, Record<string, string>]);
  const page = await ctx.newPage();
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(page.url() + ' → ' + e.message));
  (page as any).__errors = errors;
  return { ctx, page };
}

type Fixtures = {
  pageOpts: PageOpts;
  spawn: (opts?: PageOpts) => Promise<Page>;
  page: Page;
  context: BrowserContext;
};

export const test = base.extend<Fixtures>({
  pageOpts: [{ signedIn: true }, { option: true }],
  spawn: async ({}, use, testInfo) => {
    const browsers: Browser[] = [];
    const pages: Page[] = [];
    await use(async (opts = {}) => {
      const b = await launch();
      browsers.push(b);
      const { page } = await makePage(b, opts);
      pages.push(page);
      return page;
    });
    if (testInfo.status !== testInfo.expectedStatus) {
      for (let i = 0; i < pages.length; i++) {
        const shot = await pages[i].screenshot({ fullPage: false }).catch(() => null);
        if (shot) await testInfo.attach('extra-page-' + i, { body: shot, contentType: 'image/png' });
      }
    }
    for (const b of browsers) await b.close().catch(() => {});
    const errors = pages.flatMap((p) => (p as any).__errors as string[]);
    expect(errors, 'uncaught errors in the page').toEqual([]);
  },
  context: async ({ pageOpts }, use, testInfo) => {
    const b = await launch();
    const { ctx } = await makePage(b, pageOpts);
    await use(ctx);
    if (testInfo.status !== testInfo.expectedStatus) {
      const p = ctx.pages()[0];
      const shot = p ? await p.screenshot({ fullPage: false }).catch(() => null) : null;
      if (shot) await testInfo.attach('failure', { body: shot, contentType: 'image/png' });
    }
    const errors = ctx.pages().flatMap((p) => ((p as any).__errors as string[]) || []);
    await b.close().catch(() => {});
    expect(errors, 'uncaught errors in the page').toEqual([]);
  },
  page: async ({ context }, use) => {
    await use(context.pages()[0]);
  },
});

export { expect };
