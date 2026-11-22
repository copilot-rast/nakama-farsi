import fs from 'node:fs';
import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { base } from '../../lib/api';
import { restoreConfig } from '../../lib/db';
import { FIXTURES } from '../../lib/env';
import { inkOffsets, type Ink } from '../../lib/ink';
import type { Page } from '@playwright/test';

const LOGO = 'https://e2e-shell.invalid/logo.svg';
const PP = { name: { fa: 'گشت پلیس 3', en: 'Police Patrol 3' }, shortName: 'PP3', logoUrl: LOGO, logoDarkUrl: '', markUrl: '', accent: '#6b59ee', welcome: {} };
const BAR = '#main > div.sticky:first-child';

test.describe.configure({ mode: 'serial' });
test.afterEach(() => restoreConfig());

async function visit(page: Page, hash = '#/') {
  await page.route(LOGO, (r) => r.fulfill({ status: 200, contentType: 'image/svg+xml', body: fs.readFileSync(path.join(FIXTURES, 'mark.svg')) }));
  await page.goto('/' + hash);
  await page.locator('.x-drawer a[href="#/players"]').first().waitFor({ state: 'attached', timeout: 20_000 });
  await page.locator('.pl-tb-end').waitFor();
  await page.waitForFunction(() => !document.querySelector('#main .animate-pulse'));
  await page.evaluate(() => document.fonts.ready);
  await page.waitForTimeout(300);
}

type Box = { name: string; h: number; w: number; cy: number; top: number; icon: boolean };

async function controls(page: Page): Promise<{ boxes: Box[]; title: number | null; scroll: boolean }> {
  return page.evaluate((bar) => {
    const root = document.querySelector(bar)!;
    const sel = '.pl-tb-start .pl-ctl, .pl-tb-end .pl-ctl, .pl-tb-end .pl-seg-lang, .economy-ps, .justify-end > .x-popover > button.x-button';
    const boxes = [...root.querySelectorAll<HTMLElement>(sel)].filter((e) => e.getClientRects().length && getComputedStyle(e).visibility !== 'hidden').map((e) => {
      const r = e.getBoundingClientRect();
      const name = ['pl-handoff', 'pl-sdk', 'pl-up', 'pl-search', 'pl-status', 'pl-seg-lang', 'pl-theme', 'pl-menu', 'pl-side-toggle', 'economy-ps', 'x-button'].find((c) => e.classList.contains(c)) || e.tagName;
      return { name, h: r.height, w: r.width, cy: r.top + r.height / 2, top: r.top, icon: e.classList.contains('pl-ctl-icon') || !e.classList.contains('pl-ctl') && !e.classList.contains('pl-seg-lang') };
    });
    const t = root.querySelector('.pl-crumb-current');
    const tr = t ? t.getBoundingClientRect() : null;
    return { boxes, title: tr ? tr.top + tr.height / 2 : null, scroll: document.documentElement.scrollWidth > innerWidth };
  }, BAR);
}

function expectEven(m: { boxes: Box[]; title: number | null; scroll: boolean }, h: number, min: number) {
  expect(m.boxes.length, JSON.stringify(m.boxes)).toBeGreaterThanOrEqual(min);
  for (const b of m.boxes) {
    expect(b.h, b.name).toBeCloseTo(h, 0);
    if (b.icon) expect(b.w, b.name).toBeCloseTo(h, 0);
    expect(Math.abs(b.cy - m.boxes[0].cy), b.name).toBeLessThanOrEqual(0.5);
  }
  expect(m.title).not.toBeNull();
  expect(Math.abs(m.title! - m.boxes[0].cy)).toBeLessThanOrEqual(1.5);
  expect(m.scroll).toBe(false);
}

function expectCentred(ink: Ink[], label: string) {
  console.log(label, JSON.stringify(ink));
  expect(ink.length, label).toBeGreaterThanOrEqual(5);
  for (const i of ink) {
    expect(Math.abs(i.box), `${label} ${i.name} «${i.text}» ink vs box`).toBeLessThanOrEqual(1);
    if (i.dot !== null) expect(Math.abs(i.dot), `${label} ${i.name} «${i.text}» ink vs dot`).toBeLessThanOrEqual(1);
  }
}

test.describe('shell controls', () => {
  test('top bar controls share one height; brand name and app icon in the sidebar head; search matches the controls', async ({ spawn }) => {
    await base.put('/brand', PP);
    const page = await spawn({ signedIn: true, mode: 'dark' });
    await visit(page);
    const m = await controls(page);
    expectEven(m, 32, 6);
    const names = m.boxes.map((b) => b.name);
    for (const n of ['pl-sdk', 'pl-status', 'pl-seg-lang', 'pl-theme', 'economy-ps', 'x-button']) expect(names).toContain(n);
    const radii = await page.locator(`${BAR} .pl-tb-end .pl-ctl:visible`).evaluateAll((els) => [...new Set(els.map((e) => getComputedStyle(e).borderTopLeftRadius))]);
    expect(radii).toEqual(['6px']);
    const icons = await page.locator(`${BAR} :is(.pl-tb-end .pl-ctl:visible, .economy-ps, .x-popover > button.x-button) svg`).evaluateAll((els) => [...new Set(els.map((e) => Math.round(e.getBoundingClientRect().width)))]);
    expect(icons).toEqual([16]);
    const sep = await page.locator(`${BAR} .pl-tb-sep`).boundingBox();
    const seg = await page.locator(`${BAR} .pl-seg-lang`).boundingBox();
    expect(sep).not.toBeNull();
    expect(sep!.x + sep!.width).toBeLessThanOrEqual(seg!.x + seg!.width + 40);

    const ws = page.locator('.pl-side .pl-ws');
    await expect(ws.locator('.pl-ws-name b')).toHaveText('گشت پلیس ۳');
    await expect(ws.locator('.pl-ws-name span')).toHaveText('کنسول مدیریت');
    const icon = ws.locator('.pl-app-icon');
    await expect(icon.locator('img')).toBeVisible();
    const ib = await icon.boundingBox();
    expect(ib!.width).toBeCloseTo(32, 0);
    expect(ib!.height).toBeCloseTo(32, 0);
    expect(parseFloat(await icon.evaluate((e) => getComputedStyle(e).borderTopLeftRadius))).toBeGreaterThanOrEqual(6);
    const search = await page.locator('.pl-side-search').boundingBox();
    expect(search!.height).toBeCloseTo(32, 0);
    expect(await page.locator('.pl-side-search').evaluate((e) => getComputedStyle(e).borderTopLeftRadius)).toBe('6px');
    await expect(page.locator('.pl-rail .pl-brand')).toHaveCount(0);

    await page.locator('.pl-seg-lang [data-locale="en"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
    await expect(ws.locator('.pl-ws-name b')).toHaveText('Police Patrol 3');
    await expect(ws.locator('.pl-ws-name span')).toHaveText('Admin console');
    await page.waitForTimeout(400);
    expectEven(await controls(page), 32, 6);
    await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  });

  for (const mode of ['light', 'dark'] as const) {
    test(`labels sit on the optical middle of their controls and the title in ${mode}, fa and en`, async ({ spawn }) => {
      await base.put('/brand', PP);
      const page = await spawn({ signedIn: true, mode, scale: 2 });
      await visit(page);
      expectCentred(await inkOffsets(page, BAR), `fa-${mode}`);
      await page.locator('.pl-seg-lang [data-locale="en"]').click();
      await expect(page.locator('html')).toHaveAttribute('dir', 'ltr');
      await page.evaluate(() => document.fonts.ready);
      await page.waitForTimeout(400);
      expectCentred(await inkOffsets(page, BAR), `en-${mode}`);
      await page.locator('.pl-seg-lang [data-locale="fa-IR"]').click();
      await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
    });
  }

  for (const [locale, mode] of [['fa-IR', 'light'], ['en', 'dark']] as const) {
    test(`charts use the console font and text colour (${locale}, ${mode})`, async ({ spawn }) => {
      const page = await spawn({ signedIn: true, mode, locale });
      await visit(page);
      const chart = page.locator('#main [_echarts_instance_]').first();
      await expect(chart.locator('text').first()).toBeAttached();
      const native = await chart.evaluate((e) => {
        const tone = (v: string) => {
          const probe = document.createElement('span');
          probe.style.color = `var(${v})`;
          document.body.appendChild(probe);
          const c = getComputedStyle(probe).color;
          probe.remove();
          return c;
        };
        const texts = [...e.querySelectorAll('text')];
        return { body: getComputedStyle(document.body).fontFamily, axis: tone('--ink-3'), title: tone('--ink-2'), fonts: [...new Set(texts.map((t) => getComputedStyle(t).fontFamily))], fills: [...new Set(texts.map((t) => getComputedStyle(t).fill))] };
      });
      expect(native.body).toContain('Yekan Bakh');
      expect(native.fonts).toEqual([native.body]);
      expect(native.fills).toContain(native.axis);
      expect(native.fills.filter((f) => f !== native.axis && f !== native.title)).toEqual([]);
      await page.goto('/#/system/insights');
      const svg = page.locator('[data-in-chart="players"] svg');
      await expect(svg).toBeVisible();
      const overlay = await svg.evaluate((e) => [...new Set([...e.querySelectorAll('text')].map((t) => getComputedStyle(t).fontFamily))]);
      expect(overlay).toEqual([native.body]);
    });
  }

  test('at 1024px the labelled buttons fold to icons and nothing wraps', async ({ spawn }) => {
    await base.put('/brand', PP);
    const page = await spawn({ signedIn: true, width: 1024, height: 768 });
    await visit(page);
    const m = await controls(page);
    expectEven(m, 32, 6);
    const sdk = m.boxes.find((b) => b.name === 'pl-sdk')!;
    expect(sdk.w).toBeCloseTo(32, 0);
    await expect(page.locator(`${BAR} .pl-sdk-label:visible`)).toHaveCount(0);
    const tops = new Set(m.boxes.map((b) => Math.round(b.top)));
    expect(tops.size).toBe(1);
  });

  test('the sidebar collapses to the rail with the app icon, stays collapsed after a reload, and opens again', async ({ spawn }) => {
    await base.put('/brand', PP);
    const page = await spawn({ signedIn: true });
    await visit(page, '#/players');
    await page.locator('.pl-side-head .pl-side-toggle').click();
    await expect(page.locator('.pl-side')).toBeHidden();
    await expect(page.locator('.pl-rail .pl-brand .pl-app-icon img')).toBeVisible();
    const open = page.locator(`${BAR} .pl-tb-start .pl-side-toggle`);
    await expect(open).toBeVisible();
    await expect(open).toBeFocused();
    await expect(page.locator(`${BAR} .pl-tb-end .pl-search`)).toBeVisible();
    expectEven(await controls(page), 32, 8);
    const main = await page.locator('#main').evaluate((e) => parseFloat(getComputedStyle(e).paddingInlineStart));
    expect(main).toBe(60);
    await page.reload();
    await page.locator('.pl-tb-end').waitFor();
    await expect(page.locator('.pl-side')).toBeHidden();
    await page.locator(`${BAR} .pl-tb-start .pl-side-toggle`).click();
    await expect(page.locator('.pl-side')).toBeVisible();
    await expect(page.locator('.pl-rail .pl-brand')).toHaveCount(0);
    await expect(page.locator(`${BAR} .pl-tb-start .pl-side-toggle`)).toBeHidden();
    expect(await page.evaluate(() => localStorage.getItem('nakama.plane.compact'))).toBeNull();
  });

  test('mobile: 36px controls, no overflow; the menu sheet holds the Unity SDK link, language and theme', async ({ spawn }) => {
    await base.put('/brand', PP);
    const page = await spawn({ signedIn: true, mobile: true });
    await visit(page);
    const m = await controls(page);
    expectEven(m, 36, 4);
    await expect(page.locator(`${BAR} .pl-tb-end .pl-sdk:visible`)).toHaveCount(0);
    await expect(page.locator(`${BAR} .pl-tb-end .pl-seg-lang`)).toBeHidden();
    await page.locator('.pl-menu').click();
    const tools = page.locator('.pl-side .pl-side-tools');
    await expect(tools.locator('[data-pl-tool="sdk"]')).toBeVisible();
    await expect(tools.locator('[data-pl-tool="handoff"]')).toBeVisible();
    await expect(tools.locator('.pl-seg-lang [data-locale="fa-IR"]')).toHaveAttribute('aria-checked', 'true');
    const th = await tools.locator('[data-pl-tool="theme"]').boundingBox();
    const sg = await tools.locator('.pl-seg-lang').boundingBox();
    expect(th!.height).toBeCloseTo(36, 0);
    expect(sg!.height).toBeCloseTo(36, 0);
    await tools.locator('[data-pl-tool="theme"]').click();
    await expect(page.locator('html')).toHaveClass(/\bdark\b/);
    await page.locator('.pl-menu').click({ force: true }).catch(() => null);
    await page.keyboard.press('Escape');
  });

  test('no brand: the sidebar head shows the neutral icon and «کنسول مدیریت»', async ({ spawn }) => {
    await base.put('/brand', {});
    const page = await spawn({ signedIn: true });
    await visit(page);
    const ws = page.locator('.pl-side .pl-ws');
    await expect(ws).toHaveText('کنسول مدیریت');
    await expect(ws.locator('.pl-app-icon')).toHaveAttribute('data-empty', 'neutral');
    await expect(ws.locator('img')).toHaveCount(0);
  });
});
