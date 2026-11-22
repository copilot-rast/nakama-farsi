import type { Page } from '@playwright/test';
export type Ink = { name: string; text: string; box: number; dot: number | null };

async function inkY(page: Page, it: { x: number; w: number; top: number; height: number }): Promise<number> {
  const png = await page.screenshot({ clip: { x: Math.floor(it.x), y: it.top, width: Math.ceil(it.w), height: it.height } });
  const rows = await page.evaluate(async (b64) => {
    const img = new Image();
    img.src = 'data:image/png;base64,' + b64;
    await img.decode();
    const cv = document.createElement('canvas');
    cv.width = img.width;
    cv.height = img.height;
    const g = cv.getContext('2d')!;
    g.drawImage(img, 0, 0);
    const d = g.getImageData(0, 0, cv.width, cv.height).data;
    const count = new Map<number, number>();
    for (let k = 0; k < d.length; k += 4) {
      const c = (d[k] << 16) | (d[k + 1] << 8) | d[k + 2];
      count.set(c, (count.get(c) || 0) + 1);
    }
    const bg = [...count.entries()].sort((a, b) => b[1] - a[1])[0][0];
    const br = bg >> 16, bgG = (bg >> 8) & 255, bb = bg & 255;
    let top = -1;
    let bottom = -1;
    for (let y = 0; y < cv.height; y++) {
      for (let x = 0; x < cv.width; x++) {
        const k = (y * cv.width + x) * 4;
        if (Math.abs(d[k] - br) + Math.abs(d[k + 1] - bgG) + Math.abs(d[k + 2] - bb) > 90) {
          if (top < 0) top = y;
          bottom = y + 1;
        }
      }
    }
    return { y: top < 0 ? -1 : (top + bottom) / 2, scale: cv.height };
  }, png.toString('base64'));
  return it.top + rows.y * (it.height / rows.scale);
}

export async function inkOffsets(page: Page, bar: string): Promise<Ink[]> {
  const items = await page.evaluate((bar) => {
    const root = document.querySelector(bar)!;
    const sel = '.pl-tb-end .pl-status, .pl-tb-end .pl-sdk, .pl-tb-end .pl-handoff, .pl-tb-end .pl-seg-lang .pl-seg-item, .pl-tb-start .pl-crumb-current';
    return [...root.querySelectorAll<HTMLElement>(sel)].filter((e) => e.getClientRects().length && getComputedStyle(e).visibility !== 'hidden').map((e) => {
      const walk = document.createTreeWalker(e, NodeFilter.SHOW_TEXT);
      let n: Node | null;
      let tr: DOMRect | null = null;
      let text = '';
      while ((n = walk.nextNode())) {
        const r = document.createRange();
        r.selectNodeContents(n);
        const b = r.getBoundingClientRect();
        if (!n.textContent!.trim() || !b.width || !(n.parentElement as HTMLElement).getClientRects().length) continue;
        tr = b;
        text = n.textContent!.trim();
        break;
      }
      const r = e.getBoundingClientRect();
      const ctl = e.closest('.pl-ctl, .pl-seg-lang') as HTMLElement | null;
      const cr = ctl ? ctl.getBoundingClientRect() : null;
      const dot = e.querySelector('.pl-dot');
      const dr = dot ? dot.getBoundingClientRect() : null;
      const name = e.classList.contains('pl-crumb-current') ? 'title' : e.classList.contains('pl-seg-item') ? 'lang-' + (e.getAttribute('data-locale') || '') : ['pl-status', 'pl-handoff', 'pl-sdk'].find((c) => e.classList.contains(c))!;
      const top = name === 'title' ? r.top : r.top + 2;
      const height = name === 'title' ? r.height : r.height - 4;
      return { name, text, x: tr ? tr.left : 0, w: tr ? tr.width : 0, top, height, mid: cr ? cr.top + cr.height / 2 : r.top + r.height / 2, dot: dr ? dr.top + dr.height / 2 : null };
    }).filter((i) => i.w > 0);
  }, bar);
  const ctlMid = items.find((i) => i.name === 'pl-status')?.mid ?? null;
  const out: Ink[] = [];
  for (const it of items) {
    const ink = await inkY(page, it);
    const mid = it.name === 'title' ? ctlMid ?? it.mid : it.mid;
    out.push({ name: it.name, text: it.text, box: +(ink - mid).toFixed(2), dot: it.dot === null ? null : +(ink - it.dot).toFixed(2) });
  }
  return out;
}

export async function tagOffsets(page: Page, sel: string): Promise<Ink[]> {
  const items = await page.evaluate((sel) => {
    return [...document.querySelectorAll<HTMLElement>(sel)].filter((e) => e.getClientRects().length && getComputedStyle(e).visibility !== 'hidden').map((e) => {
      const walk = document.createTreeWalker(e, NodeFilter.SHOW_TEXT);
      let n: Node | null;
      let tr: DOMRect | null = null;
      let text = '';
      while ((n = walk.nextNode())) {
        const r = document.createRange();
        r.selectNodeContents(n);
        const b = r.getBoundingClientRect();
        if (!n.textContent!.trim() || !b.width) continue;
        tr = b;
        text = n.textContent!.trim();
        break;
      }
      const r = e.getBoundingClientRect();
      const host = e.closest('.x-badge');
      const dot = host ? host.querySelector('.absolute.rounded-full') : null;
      const dr = dot ? dot.getBoundingClientRect() : null;
      return { name: e.className.includes('eco-tag') ? 'eco-tag' : host ? 'x-badge' : 'x-tag', text, x: tr ? tr.left : 0, w: tr ? Math.min(tr.width, r.right - tr.left) : 0, top: r.top + 1, height: r.height - 2, mid: r.top + r.height / 2, dot: dr ? dr.top + dr.height / 2 : null };
    }).filter((i) => i.w > 0);
  }, sel);
  const out: Ink[] = [];
  for (const it of items) {
    const ink = await inkY(page, it);
    out.push({ name: it.name, text: it.text, box: +(ink - it.mid).toFixed(2), dot: it.dot === null ? null : +(ink - it.dot).toFixed(2) });
  }
  return out;
}
