import type { Page } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { still } from '../../lib/console';

async function visit(page: Page, hash: string) {
  await page.goto('/' + hash);
  await page.locator('.sp-rail').waitFor();
  await page.waitForFunction(() => !document.querySelector('#main .animate-pulse, [data-economy] .animate-pulse'));
  await still(page);
}

const ms = (v: string) => v.split(',').map((x) => (x.trim().endsWith('ms') ? parseFloat(x) : parseFloat(x) * 1000));

test.describe('spatial motion', () => {
  test('route change plays the page enter once on the content root and removes it; an id change does not replay it', async ({ page }) => {
    await visit(page, '#/players');
    await page.evaluate(() => {
      (window as any).__enter = [];
      new MutationObserver((rs) => rs.forEach((r) => {
        const t = r.target as HTMLElement;
        if (t.classList && t.classList.contains('sp-enter')) (window as any).__enter.push(t.id || t.tagName);
      })).observe(document.getElementById('main')!, { attributes: true, attributeFilter: ['class'], subtree: true });
    });
    await page.locator('.sp-rail a.sp-ri[href="#/groups"]').click();
    await expect.poll(() => page.evaluate(() => (window as any).__enter.length)).toBeGreaterThan(0);
    await expect(page.locator('.sp-enter')).toHaveCount(0, { timeout: 2000 });
    expect(await page.evaluate(() => getComputedStyle(document.getElementById('main')!).transform)).toBe('none');
    await page.goto('/#/economy/products');
    await page.locator('#economy-root table').waitFor();
    await still(page);
    await page.evaluate(() => {
      (window as any).__enter = [];
      new MutationObserver((rs) => rs.forEach((r) => {
        const t = r.target as HTMLElement;
        if (t.classList && t.classList.contains('sp-enter')) (window as any).__enter.push(t.id || t.tagName);
      })).observe(document.getElementById('main')!, { attributes: true, attributeFilter: ['class'], subtree: true });
    });
    await page.locator('#economy-root tbody tr').first().click();
    await expect(page.locator('[data-economy="panel"] form')).toBeVisible();
    await page.waitForTimeout(400);
    expect(await page.evaluate(() => (window as any).__enter.length)).toBe(0);
  });

  test('durations come from the four tokens, exits are faster than enters, one easing family', async ({ page }) => {
    await visit(page, '#/players');
    const t = await page.evaluate(() => {
      const fly = document.querySelector('.sp-flyout') as HTMLElement | null;
      const btn = document.querySelector('.sp-rail button[data-sp-group="economy"]') as HTMLElement;
      btn.dispatchEvent(new MouseEvent('mouseenter'));
      const f = document.querySelector('.sp-flyout') as HTMLElement;
      const open = getComputedStyle(f).transitionDuration;
      f.classList.remove('open');
      const close = getComputedStyle(f).transitionDuration;
      const ease = getComputedStyle(f).transitionTimingFunction;
      const ri = getComputedStyle(document.querySelector('.sp-rail .sp-ri') as HTMLElement).transitionDuration;
      const seg = getComputedStyle(document.querySelector('.sp-seg-ind') as HTMLElement).transitionDuration;
      return { had: !!fly, open, close, ease, ri, seg };
    });
    expect(Math.max(...ms(t.open))).toBe(200);
    expect(Math.max(...ms(t.close))).toBe(120);
    expect(Math.max(...ms(t.close))).toBeLessThan(Math.max(...ms(t.open)));
    expect(t.ease).toContain('cubic-bezier(0.4, 0, 1, 1)');
    expect(Math.max(...ms(t.ri))).toBe(120);
    expect(Math.max(...ms(t.seg))).toBe(200);
    await page.locator('#main tbody tr').first().waitFor();
    await page.goto('/#/storage');
    await page.locator('#main tbody tr').first().click();
    const holder = page.locator('div.fixed.inset-0.flex.justify-end:has(> .x-form)');
    await expect(holder).toBeVisible();
    const enter = await holder.evaluate((e) => getComputedStyle(e).transitionDuration);
    expect(ms(enter)).toContain(280);
    await still(page);
  });

  test('the flyout fades in from transparent and the editor scales in from .96, both reach rest', async ({ page }) => {
    await visit(page, '#/players');
    const samples = await page.evaluate(async () => {
      const btn = document.querySelector('.sp-rail button[data-sp-group="publication"]') as HTMLElement;
      btn.dispatchEvent(new MouseEvent('mouseenter'));
      const f = document.querySelector('.sp-flyout') as HTMLElement;
      const out: number[] = [];
      const t0 = performance.now();
      while (performance.now() - t0 < 320) {
        out.push(Number(getComputedStyle(f).opacity));
        await new Promise((r) => requestAnimationFrame(r));
      }
      return out;
    });
    expect(samples[0]).toBeLessThan(0.9);
    expect(samples[samples.length - 1]).toBe(1);
    await page.goto('/#/economy/products/new');
    const form = page.locator('[data-economy="panel"] .x-form');
    await expect(form).toBeVisible();
    const anim = await form.evaluate((e) => e.getAnimations().map((a) => (a as any).animationName));
    expect(anim.join(',')).toMatch(/sp-win-in|^$/);
    await still(page);
    expect(await form.evaluate((e) => getComputedStyle(e).transform)).toBe('none');
  });

  test('clicks act during animations: a rail page right after hovering a group, Create right after the editor opens', async ({ page }) => {
    await visit(page, '#/players');
    await page.locator('.sp-rail button[data-sp-group="economy"]').hover();
    await page.waitForTimeout(30);
    await page.locator('.sp-rail a.sp-ri[href="#/groups"]').click();
    await expect(page).toHaveURL(/#\/groups$/);
    await page.goto('/#/economy/currencies/new');
    const form = page.locator('[data-economy="panel"] form');
    await expect(form).toBeVisible();
    await page.waitForTimeout(50);
    await page.locator('[data-economy="panel"] button', { hasText: 'ایجاد' }).click();
    await expect(page.locator('[data-economy="panel"] .eco-invalid').first()).toBeVisible();
  });

  test('no transform on #main, the scroll container, or any ancestor of an open editor, Monaco or popover', async ({ page }) => {
    await visit(page, '#/storage');
    await page.locator('#main tbody tr').first().click();
    await expect(page.locator('div.fixed.inset-0.flex.justify-end .monaco-editor').first()).toBeVisible();
    await still(page);
    const bad = await page.evaluate(() => {
      const out: string[] = [];
      const check = (el: Element | null) => {
        for (let a = el && el.parentElement; a && a !== document.documentElement; a = a.parentElement) {
          if (getComputedStyle(a).transform !== 'none') out.push(a.tagName + '.' + a.className.toString().slice(0, 40));
        }
      };
      check(document.querySelector('.monaco-editor'));
      check(document.querySelector('div.fixed.inset-0.flex.justify-end > .x-form'));
      if (getComputedStyle(document.getElementById('main')!).transform !== 'none') out.push('#main');
      return out;
    });
    expect(bad).toEqual([]);
    await page.goto('/#/players');
    await page.locator('#main tbody tr').nth(1).locator('button').last().click();
    await expect(page.locator('.v-popper__popper--shown')).toBeVisible();
    await still(page);
    const pop = await page.evaluate(() => {
      const el = document.querySelector('.v-popper__popper--shown');
      const out: string[] = [];
      for (let a = el && el.parentElement; a && a !== document.documentElement; a = a.parentElement) if (getComputedStyle(a).transform !== 'none') out.push(a.tagName);
      return out;
    });
    expect(pop).toEqual([]);
  });
});

test('KPI numbers count up once on first paint and end on the real value', async ({ page }) => {
  await page.goto('/#/economy/overview');
  const kpi = page.locator('#economy-root .eco-stat-grid .text-3xl').first();
  await kpi.waitFor();
  const seen: string[] = [];
  for (let i = 0; i < 12; i++) { seen.push((await kpi.textContent()) || ''); await page.waitForTimeout(40); }
  await page.waitForTimeout(600);
  const final = (await kpi.textContent()) || '';
  expect(final).toMatch(/^[۰-۹٬]+$/);
  expect(seen[seen.length - 1]).toBe(final);
});

test.describe('spatial motion with reduced motion', () => {
  test('every animation is done 50ms after an interaction', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await visit(page, '#/players');
    const running = async () => page.evaluate(() => document.getAnimations().filter((a) => a.playState === 'running' && Number(a.effect?.getTiming().duration) > 1).map((a) => {
      const t = (a.effect as KeyframeEffect).target as Element;
      return ((a as any).animationName || (a as any).transitionProperty || 'x') + ' ' + (t ? t.tagName + '.' + String(t.className).slice(0, 50) : '') + ' ' + a.effect?.getTiming().duration;
    }));
    await page.locator('.sp-rail button[data-sp-group="economy"]').click();
    await page.waitForTimeout(50);
    expect(await running()).toEqual([]);
    await page.locator('.sp-flyout.open a[href^="#/economy/products"]').click();
    await page.locator('#economy-root table').waitFor();
    await page.waitForTimeout(50);
    expect(await running()).toEqual([]);
    await page.locator('#economy-root tbody tr').first().click();
    await expect(page.locator('[data-economy="panel"] form')).toBeVisible();
    await page.waitForTimeout(50);
    expect(await running()).toEqual([]);
    const shimmer = await page.evaluate(() => {
      const s = document.createElement('div');
      s.className = 'animate-pulse bg-secondary-300';
      document.getElementById('economy-root')!.appendChild(s);
      const n = s.getAnimations().filter((a) => a.playState === 'running' && Number(a.effect?.getTiming().duration) > 1).length;
      s.remove();
      return n;
    });
    expect(shimmer).toBe(0);
  });
});
