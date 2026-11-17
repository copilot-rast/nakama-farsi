import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import type { Page } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { noHorizontalScroll, open, panel, root } from '../../lib/console';
import { base } from '../../lib/api';

const TAG = 'e2e' + Date.now().toString(36);
const MAP = TAG + '_storm';
const SAMPLES = process.env.E2E_HEATMAP_SAMPLES || '';

function lap(points: number, player: string, header = true, shift = 0) {
  const lines = header ? ['t,x,z,player,kmh'] : [];
  for (let i = 0; i < points; i++) {
    const a = (i / points) * Math.PI * 2;
    lines.push([(i * 0.2).toFixed(1), (Math.cos(a) * 120 + shift).toFixed(2), (Math.sin(a) * 70).toFixed(2), player, (60 + (i % 80)).toString()].join(','));
  }
  return lines.join('\n') + '\n';
}

function storeZip(files: Record<string, string>) {
  const crcTable = Array.from({ length: 256 }, (_, n) => {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    return c >>> 0;
  });
  const crc = (b: Buffer) => {
    let c = 0xffffffff;
    for (const x of b) c = crcTable[(c ^ x) & 0xff] ^ (c >>> 8);
    return (c ^ 0xffffffff) >>> 0;
  };
  const locals: Buffer[] = [];
  const centrals: Buffer[] = [];
  let off = 0;
  for (const [name, text] of Object.entries(files)) {
    const data = Buffer.from(text);
    const n = Buffer.from(name);
    const sum = crc(data);
    const lh = Buffer.alloc(30);
    lh.writeUInt32LE(0x04034b50, 0); lh.writeUInt16LE(20, 4); lh.writeUInt32LE(sum, 14); lh.writeUInt32LE(data.length, 18); lh.writeUInt32LE(data.length, 22); lh.writeUInt16LE(n.length, 26);
    const ch = Buffer.alloc(46);
    ch.writeUInt32LE(0x02014b50, 0); ch.writeUInt16LE(20, 4); ch.writeUInt16LE(20, 6); ch.writeUInt32LE(sum, 16); ch.writeUInt32LE(data.length, 20); ch.writeUInt32LE(data.length, 24); ch.writeUInt16LE(n.length, 28); ch.writeUInt32LE(off, 42);
    locals.push(lh, n, data);
    centrals.push(ch, n);
    off += 30 + n.length + data.length;
  }
  const cd = Buffer.concat(centrals);
  const end = Buffer.alloc(22);
  end.writeUInt32LE(0x06054b50, 0); end.writeUInt16LE(Object.keys(files).length, 8); end.writeUInt16LE(Object.keys(files).length, 10); end.writeUInt32LE(cd.length, 12); end.writeUInt32LE(off, 16);
  return Buffer.concat([...locals, cd, end]);
}

function samples() {
  if (!SAMPLES || !fs.existsSync(SAMPLES)) return [];
  return fs.readdirSync(SAMPLES).filter((f) => /\.(csv|csv\.gz)$/i.test(f)).sort().slice(0, 3).map((f) => ({ name: f, mimeType: 'text/csv', buffer: fs.readFileSync(path.join(SAMPLES, f)) }));
}

async function openUpload(page: Page) {
  await root(page).locator('[data-in-hm-upload]').first().click();
  await expect(panel(page).locator('[data-in-upload]')).toBeVisible();
  return panel(page);
}

test.describe.configure({ mode: 'serial' });

test.describe('heatmap CSV upload from the console', () => {
  test.afterAll(async () => {
    await base.del('/insights/heatmaps/' + MAP);
  });

  test('a new map from plain, gzip and zip files with a per-file report', async ({ page }) => {
    await open(page, '#/system/insights/heatmaps');
    const p = await openUpload(page);
    await p.locator('[data-in-up-target] [data-value="new"]').click();
    await p.locator('[data-in-up-key]').fill(MAP);
    await p.locator('[data-in-up-name]').fill('طوفان جاده');
    await p.locator('[data-in-up-help] summary').click();
    await expect(p.locator('[data-in-up-help]')).toContainText('12.5,104.2,-33.7,Sara,86');
    const bad = lap(100, 'Sara') + '1,2\nfoo,bar,baz,me,1\n3,4,5,me,99999\n';
    const real = samples();
    await p.locator('[data-in-up-input]').setInputFiles([
      { name: 'race-1.csv', mimeType: 'text/csv', buffer: Buffer.from(bad) },
      { name: 'race-2.csv.gz', mimeType: 'application/gzip', buffer: zlib.gzipSync(lap(80, 'Reza', false, 4)) },
      { name: 'cup.zip', mimeType: 'application/zip', buffer: storeZip({ 'heat-a.csv': lap(40, 'Ali'), 'heat-b.csv': lap(60, 'Mina', true, 8) }) },
      ...real,
    ]);
    await expect(p.locator('[data-in-up-file]')).toHaveCount(3 + real.length);
    await p.locator('[data-in-up-send]').click();
    const results = p.locator('[data-in-up-results]');
    await expect(p.locator('[data-in-up-view]')).toBeVisible({ timeout: 60_000 });
    await expect(results.locator('[data-in-up-status="stored"]')).toHaveCount(4 + real.length);
    const row1 = results.locator('tr', { hasText: 'race-1.csv' });
    await expect(row1).toContainText('۱۰۳');
    await expect(row1).toContainText('ستون کم: ۱');
    await expect(row1).toContainText('عدد نادرست: ۱');
    await expect(row1).toContainText('بیرون از محدوده: ۱');
    await expect(row1.locator('[data-in-up-points]')).toHaveText('۱۰۰');
    await expect(results.locator('tr', { hasText: 'cup.zip/heat-b.csv' }).locator('[data-in-up-points]')).toHaveText('۶۰');
    const map = (await base.get('/insights/heatmaps')).items.find((m: any) => m.key === MAP);
    expect(map).toMatchObject({ name: 'طوفان جاده', uploads: 4 + real.length });
    const total = map.points;
    expect(total).toBeGreaterThanOrEqual(280);
    await p.locator('[data-in-up-view]').click();
    await expect(panel(page)).toHaveCount(0);
    await expect(root(page).locator('[data-in-heat="' + MAP + '"] canvas')).toBeVisible();
    await expect(root(page).locator('[data-in-kpi="points"]')).toContainText(total.toLocaleString('fa-IR'));
    const recent = root(page).locator('[data-in-hm-uploads]');
    await expect(recent.locator('[data-in-src="console"]').first()).toContainText('کنسول');
    await expect(recent).toContainText('heat-a');
    await recent.locator('[data-in-trace-btn]').first().click();
    await expect(root(page).locator('[data-in-trace]')).toBeVisible();
  });

  test('the same file counts once, replace starts the map over', async ({ page }) => {
    await open(page, '#/system/insights/heatmaps?map=' + MAP);
    let p = await openUpload(page);
    await expect(p.locator('[data-in-up-map]')).toHaveValue(MAP);
    await p.locator('[data-in-up-input]').setInputFiles([{ name: 'again.csv', mimeType: 'text/csv', buffer: Buffer.from(lap(40, 'Ali')) }]);
    await p.locator('[data-in-up-race] [data-value="label"]').click();
    await p.locator('[data-in-up-label]').fill('فینال');
    await p.locator('[data-in-up-send]').click();
    await expect(p.locator('[data-in-up-results] [data-in-up-status="duplicate"]')).toBeVisible({ timeout: 30_000 });
    await expect(p.locator('[data-in-up-results]')).toContainText('تکراری، شمرده نشد');
    await p.locator('[data-in-up-view]').click();
    p = await openUpload(page);
    await p.locator('[data-in-up-mode] [data-value="replace"]').click();
    await expect(p).toContainText('پاک می‌شوند');
    await p.locator('[data-in-up-player] [data-value="fixed"]').click();
    await p.locator('[data-in-up-player-name]').fill('Tester');
    await p.locator('[data-in-up-input]').setInputFiles([{ name: 'only.csv', mimeType: 'text/csv', buffer: Buffer.from(lap(50, 'x', true, 2)) }]);
    await p.locator('[data-in-up-send]').click();
    await expect(p.locator('[data-in-up-results] [data-in-up-status="stored"]')).toBeVisible({ timeout: 30_000 });
    const map = (await base.get('/insights/heatmaps')).items.find((m: any) => m.key === MAP);
    expect(map).toMatchObject({ uploads: 1, points: 50 });
    const ups = await base.get('/insights/heatmaps/' + MAP + '/uploads');
    expect(ups.items[0]).toMatchObject({ source: 'console', fileName: 'only.csv', label: 'only', points: 50 });
    const tr = await base.get('/insights/heatmaps/uploads/' + ups.items[0].id + '/trace');
    expect(tr.players.map((x: any) => x.player)).toEqual(['Tester']);
    await p.locator('[data-in-up-view]').click();
    await expect(root(page).locator('[data-in-kpi="points"]')).toContainText('۵۰');
  });

  test('English, dark, 390px', async ({ spawn }) => {
    const m = await spawn({ signedIn: true, mobile: true, locale: 'en', mode: 'dark', width: 390, height: 844 });
    await open(m, '#/system/insights/heatmaps?map=' + MAP);
    await expect(root(m).locator('[data-in-heat="' + MAP + '"] canvas')).toBeVisible();
    await noHorizontalScroll(m);
    const p = await openUpload(m);
    await expect(p).toContainText('Upload heatmap CSV');
    await expect(p.locator('[data-in-up-drop]')).toContainText('Drag and drop CSV files here');
    await p.locator('[data-in-up-input]').setInputFiles([{ name: 'mobile-lap-with-a-rather-long-file-name.csv', mimeType: 'text/csv', buffer: Buffer.from(lap(30, 'M', true, 12) + 'x,y\n') }]);
    await p.locator('[data-in-up-send]').click();
    const row = p.locator('[data-in-up-results] tr', { hasText: 'mobile-lap' });
    await expect(row.locator('[data-in-up-status="stored"]')).toBeVisible({ timeout: 30_000 });
    await expect(row).toContainText('too few columns: 1');
    await expect(row.locator('[data-in-up-points]')).toHaveText('30');
    const over = await m.evaluate(() => {
      const el = document.querySelector('[data-economy="panel"] [role="dialog"] .overflow-y-auto') as HTMLElement;
      return el ? el.scrollWidth - el.clientWidth : 999;
    });
    expect(over).toBeLessThanOrEqual(1);
    await noHorizontalScroll(m);
  });
});
