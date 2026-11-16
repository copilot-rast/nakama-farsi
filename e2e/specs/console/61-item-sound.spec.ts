import path from 'node:path';
import { expect, test } from '../../lib/fixtures';
import { FIXTURES } from '../../lib/env';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, fillTranslated, noHorizontalScroll, open, panel, pickRef, savePanel, settle, settlePanel } from '../../lib/console';

const k = uid();
const made = { types: [] as string[], users: [] as string[] };
let type: any;
let P: Player;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async () => {
  type = await base.post('/categories', { key: 'e2e_' + k + '_snd', keyType: 'item_type', name: { fa: 'صدای گل ' + k.slice(-4) }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  P = await devicePlayer('e2e-snd-' + k, ('snd' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(P.id);
});

test.afterAll(async () => {
  const ids = sql(`SELECT string_agg(quote_literal(id::text), ',') FROM economy_item WHERE key LIKE 'e2e_${k}%'`);
  if (ids) sql(`DELETE FROM economy_user_item WHERE item_id::text IN (${ids})`);
  for (const id of (ids || '').split(',').filter(Boolean)) await eco.del('/items/' + id.replace(/'/g, ''));
  for (const id of made.users) await deleteUser(id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('items: a sound for an item', () => {
  test('an item gets an uploaded sound in the editor; mp3/ogg/wav are accepted and an image is not', async ({ page }) => {
    await open(page, '#/economy/items/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill('e2e_' + k + '_goal');
    await fillTranslated(p, 'name', { fa: 'گل طلایی ' + k.slice(-4) });
    await pickRef(p, 'نوع', type.key);
    const sound = field(p, 'صدا');
    await sound.locator('input[type="file"]').setInputFiles(path.join(FIXTURES, 'goal.wav'));
    await expect(sound.locator('[data-sound-preview]')).toBeVisible();
    await savePanel(page);
    const item = (await eco.get('/items?q=' + encodeURIComponent('e2e_' + k + '_goal'))).items[0];
    expect(item.soundUrl).toMatch(/\/v2\/economy\/files\/[a-f0-9]{32}\.wav$/);
    const file = await fetch(item.soundUrl);
    expect(file.status).toBe(200);
    expect(file.headers.get('content-type')).toContain('audio/wav');

    await eco.post('/users/' + P.id + '/items', { itemId: item.id });
    const inv = await rpc(P, 'economy_inventory', {});
    expect(inv.status).toBe(200);
    const mine = (inv.data.items as any[]).find((x) => x.key === 'e2e_' + k + '_goal');
    expect(mine.soundUrl).toBe(item.soundUrl);

    await open(page, '#/economy/items/new');
    await settlePanel(page);
    const q = panel(page);
    await field(q, 'صدا').locator('input[type="file"]').setInputFiles(path.join(FIXTURES, 'logo.svg'));
    await expect(field(q, 'صدا')).toContainText('mp3');
  });

  test('the item editor fits a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await open(page, '#/economy/items/new');
    await settle(page);
    await noHorizontalScroll(page);
  });
});
