import vm from 'node:vm';
import { bundle } from '../scripts/build.mjs';

export function load(opts = {}) {
  const store = new Map(Object.entries(opts.storage || {}));
  const html = { lang: opts.lang || 'fa-IR' };
  const ctx = {
    console,
    setTimeout,
    clearTimeout,
    localStorage: { getItem: (k) => (store.has(k) ? store.get(k) : null), setItem: (k, v) => store.set(k, String(v)), removeItem: (k) => store.delete(k) },
    sessionStorage: { getItem: () => null, setItem() {}, removeItem() {} },
    document: { readyState: 'complete', documentElement: { getAttribute: (n) => (n === 'lang' ? html.lang : null) }, addEventListener() {} },
    location: { reload() { ctx.reloaded = (ctx.reloaded || 0) + 1; } },
  };
  ctx.window = ctx;
  vm.createContext(ctx);
  ctx.Intl = vm.runInContext('Intl', ctx);
  ctx.Date = vm.runInContext('Date', ctx);
  vm.runInContext(bundle().js, ctx);
  return { ND: ctx.NakamaDate, ctx, html, store };
}
