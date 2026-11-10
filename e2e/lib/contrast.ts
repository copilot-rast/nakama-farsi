import type { Page } from '@playwright/test';

export type Run = { text: string; ratio: number; size: number; at: [number, number] };

export async function measureContrast(page: Page, min = 4.5): Promise<{ count: number; min: number; fails: Run[] }> {
  const runs = await page.evaluate(() => {
    const parse = (c: string) => { const m = (c.match(/[\d.]+/g) || []).map(Number); return [m[0], m[1], m[2], m.length > 3 ? m[3] : 1]; };
    const out: any[] = [];
    const W = innerWidth, H = innerHeight;
    const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
    let n: Node | null;
    while ((n = walker.nextNode())) {
      if (!n.textContent || !n.textContent.trim()) continue;
      const el = n.parentElement;
      if (!el || el.closest('script,style,noscript,.sp-sky,[aria-hidden="true"],[inert],:disabled,[aria-disabled="true"],.x-drawer,.monaco-editor,.sr-only,[_echarts_instance_],.brand-preview,[data-sp-skip-contrast]')) continue;
      const cs = getComputedStyle(el);
      if (cs.visibility !== 'visible' || cs.display === 'none') continue;
      const col = parse(cs.color);
      let op = 1;
      for (let a: Element | null = el; a && a !== document.documentElement; a = a.parentElement) op *= Number(getComputedStyle(a).opacity);
      col[3] *= op;
      if (col[3] < 0.05) continue;
      const r = document.createRange();
      r.selectNodeContents(n);
      for (const b of Array.from(r.getClientRects())) {
        if (b.width < 3 || b.height < 6 || b.bottom <= 0 || b.top >= H || b.right <= 0 || b.left >= W) continue;
        const cx = Math.min(W - 1, Math.max(0, b.left + b.width / 2)), cy = Math.min(H - 1, Math.max(0, b.top + b.height / 2));
        let clipped = false;
        for (let a: Element | null = el; a && a !== document.documentElement; a = a.parentElement) {
          const st = getComputedStyle(a);
          if (st.overflowX === 'visible' && st.overflowY === 'visible') continue;
          const cr = a.getBoundingClientRect();
          if (cy < cr.top || cy > cr.bottom || cx < cr.left || cx > cr.right) { clipped = true; break; }
        }
        if (clipped) continue;
        const hit = document.elementFromPoint(cx, cy);
        if (!hit || !(hit === el || el.contains(hit) || hit.contains(el))) continue;
        out.push({ text: n.textContent.trim().slice(0, 40), color: col, size: parseFloat(cs.fontSize), rect: [b.left, b.top, b.width, b.height] });
      }
    }
    return out;
  });
  const hide = await page.addStyleTag({ content: '*{color:transparent!important;-webkit-text-fill-color:transparent!important;text-shadow:none!important;caret-color:transparent!important;text-decoration-color:transparent!important}input::placeholder,textarea::placeholder{color:transparent!important}' });
  await page.waitForTimeout(120);
  const png = (await page.screenshot({ type: 'png' })).toString('base64');
  await hide.evaluate((t) => t.remove());
  const measured: Run[] = await page.evaluate(async ([b64, runs]) => {
    const img = new Image();
    img.src = 'data:image/png;base64,' + b64;
    await img.decode();
    const c = document.createElement('canvas');
    c.width = img.width; c.height = img.height;
    const ctx = c.getContext('2d')!;
    ctx.drawImage(img, 0, 0);
    const data = ctx.getImageData(0, 0, c.width, c.height).data;
    const sx = img.width / innerWidth;
    const lin = (v: number) => { v /= 255; return v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4; };
    const lum = (p: number[]) => 0.2126 * lin(p[0]) + 0.7152 * lin(p[1]) + 0.0722 * lin(p[2]);
    const ratio = (a: number[], b: number[]) => { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); };
    return (runs as any[]).map((r) => {
      const [x, y, w, h] = r.rect;
      const vals: number[] = [];
      for (let i = 1; i <= 6; i++) for (let j = 1; j <= 4; j++) {
        const px = Math.floor((x + (w * i) / 7) * sx), py = Math.floor((y + (h * j) / 5) * sx);
        if (px < 0 || py < 0 || px >= c.width || py >= c.height) continue;
        const k = (py * c.width + px) * 4;
        const bg = [data[k], data[k + 1], data[k + 2]];
        const a = r.color[3];
        const fg = [0, 1, 2].map((q) => r.color[q] * a + bg[q] * (1 - a));
        vals.push(ratio(fg, bg));
      }
      vals.sort((p, q) => p - q);
      const p10 = vals.length ? vals[Math.max(0, Math.floor(vals.length * 0.1) - 1)] : 99;
      return { text: r.text, ratio: Math.round(p10 * 100) / 100, size: r.size, at: [Math.round(x), Math.round(y)] as [number, number] };
    });
  }, [png, runs] as [string, any[]]);
  const fails = measured.filter((r) => r.ratio < min).sort((a, b) => a.ratio - b.ratio);
  return { count: measured.length, min: measured.reduce((m, r) => Math.min(m, r.ratio), 99), fails };
}
