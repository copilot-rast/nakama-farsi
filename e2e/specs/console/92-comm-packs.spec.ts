import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, type Player } from '../../lib/api';
import { Chat } from '../../lib/chat';
import { sql } from '../../lib/db';
import { field, fillTranslated, open, panel, root, rows, savePanel, search, settlePanel, toasts } from '../../lib/console';

const k = uid();
const words = 'e2e_' + k + '_words';
const faces = 'e2e_' + k + '_faces';
const made = { users: [] as string[], items: [] as string[], types: [] as string[] };
const sockets: Chat[] = [];
let saved = '';
let turn = '';
let item: any;
let A: Player;
let B: Player;
let matchId = '';
let a: Chat;
let b: Chat;

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

async function socket(p: Player) {
  const c = await Chat.open(p);
  sockets.push(c);
  return c;
}

test.beforeAll(async () => {
  saved = sql(`SELECT settings::text FROM base_comm_settings WHERE id=1`);
  turn = sql(`SELECT COALESCE(turn_seconds::text,'NULL') FROM economy_game WHERE key='stargame_tictactoe'`);
  sql(`UPDATE economy_game SET turn_seconds=600 WHERE key='stargame_tictactoe'`);
  const type = await base.post('/categories', { key: 'e2e_' + k + '_cmt', keyType: 'item_type', name: { fa: 'بستهٔ پیام' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  item = await eco.post('/items', { key: 'e2e_' + k + '_cmitem', typeId: type.id, name: { fa: 'چهره‌های ویژه', en: 'Special faces' } });
  made.items.push(item.id);
  await base.post('/stickers/packs', { key: faces, kind: 'emoji', inChat: false, inMatch: true, access: 'item', itemKey: item.key, sort: -999, name: { fa: 'چهره‌ها', en: 'Faces' }, stickers: [{ key: 'angry', assetId: 'emoji.angry' }] });
  A = await devicePlayer('e2e-cm-a-' + k, ('cma' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-cm-b-' + k, ('cmb' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id);
  for (const p of [A, B]) sql(`UPDATE users SET wallet = jsonb_set(COALESCE(wallet,'{}'::jsonb), '{coin}', '500'), lang_tag='fa' WHERE id='${p.id}'`);
});

test.afterAll(async () => {
  for (const s of sockets) s.close();
  if (turn) sql(`UPDATE economy_game SET turn_seconds=${turn} WHERE key='stargame_tictactoe'`);
  if (saved) sql(`UPDATE base_comm_settings SET settings=${q(saved)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_comm_settings`);
  if (matchId) sql(`DELETE FROM base_comm_channel WHERE match_id='${matchId}'`);
  sql(`DELETE FROM base_comm_favorite WHERE user_id IN ('${A.id}','${B.id}')`);
  sql(`DELETE FROM base_sticker_pack WHERE key LIKE 'e2e_${k}%'`);
  for (const id of made.users) await deleteUser(id);
  if (made.items.length) sql(`DELETE FROM economy_user_item WHERE item_id IN (${made.items.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('communication packs: console', () => {
  test('a text pack for matches only is made in the pack editor and listed with its kind and place', async ({ page }) => {
    await open(page, '#/moderation/stickers/new');
    await settlePanel(page);
    const p = panel(page);
    await field(p, 'کلید').locator('input').fill(words);
    await fillTranslated(p, 'name', { fa: 'جمله‌ها ' + k.slice(-4), en: 'Words' });
    await field(p, 'نوع بسته').locator('select').selectOption('text');
    await p.locator('label.x-toggle', { hasText: 'در گفتگوی چت' }).click();
    await p.locator('label.x-toggle', { hasText: 'در مسابقه' }).click();
    await p.locator('[data-rows-add="stickers"]').click();
    await p.locator('[data-col="key"]').nth(0).locator('input').fill('gg');
    await p.locator('[data-col="text"]').nth(0).locator('input, textarea').first().fill('بازی خوبی بود');
    await savePanel(page);
    const pack = (await base.get('/stickers/packs?q=' + encodeURIComponent(words))).items[0];
    expect(pack).toMatchObject({ key: words, kind: 'text', inChat: false, inMatch: true, access: 'free' });
    expect(pack.stickers[0]).toMatchObject({ key: 'gg', imageUrl: '', text: { fa: 'بازی خوبی بود' } });
    await search(page, words);
    const kind = rows(page).first().locator('[data-cm-kind="text"]');
    await expect(kind).toContainText('متن');
    await expect(kind).toContainText('مسابقه');
    await expect(kind).not.toContainText('گفتگو');
  });

  test('match chat settings check the window against count and gap, then save', async ({ page }) => {
    await open(page, '#/moderation/stickers/match-chat');
    const form = root(page).locator('[data-settings-form="comms"]');
    await expect(form.locator('label.x-toggle', { hasText: 'پیام سریع بازی روشن باشد' })).toBeVisible();
    await field(form, 'فاصلهٔ دو پیام (ثانیه)').locator('input').fill('10');
    await form.locator('[data-settings-save]').click();
    await expect(form).toContainText('پنجره باید بیشتر از');
    await field(form, 'فاصلهٔ دو پیام (ثانیه)').locator('input').fill('0');
    await field(form, 'بیشترین پیام در پنجره').locator('input').fill('2');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toContainText('ذخیره');
    expect(await base.get('/comms/settings')).toMatchObject({ enabled: true, minIntervalSeconds: 0, maxCount: 2, windowSeconds: 30, lockSeconds: 30, channelHours: 2, maxFavorites: 0 });
  });
});

test.describe('communication packs: player', () => {
  test('base_comm_packs lists the free text pack owned and the item pack locked; favorites need ownership', async () => {
    const res = await rpc(A, 'base_comm_packs', { lang: 'fa' });
    expect(res.status).toBe(200);
    expect(res.data).toMatchObject({ enabled: true, cooldown: { minIntervalSeconds: 0, maxCount: 2, windowSeconds: 30, lockSeconds: 30 } });
    const mine = (res.data.packs as any[]).filter((p) => p.key === words || p.key === faces);
    expect(mine.map((p) => p.key)).toEqual([faces, words]);
    expect(mine[0]).toMatchObject({ kind: 'emoji', owned: false, starter: false, itemKey: item.key, item: { key: item.key, name: 'چهره‌های ویژه' } });
    expect(mine[0].items[0]).toMatchObject({ key: 'angry', assetId: 'emoji.angry', favorite: false });
    expect(mine[1]).toMatchObject({ kind: 'text', owned: true, starter: true });
    expect(mine[1].items[0]).toMatchObject({ ref: words + '/gg', text: 'بازی خوبی بود' });
    const locked = await rpc(A, 'base_comm_favorite', { pack: faces, item: 'angry' });
    expect(locked.status).toBe(400);
    expect(JSON.stringify(locked.data)).toContain('PACK_NOT_OWNED');
    expect((await rpc(A, 'base_comm_favorite', { pack: words, item: 'gg' })).data).toEqual({ pack: words, item: 'gg', favorite: true });
    const favs = await rpc(A, 'base_comm_favorites', {});
    expect(favs.data.items.map((x: any) => x.ref)).toEqual([words + '/gg']);
    expect((await rpc(A, 'base_stickers', {})).data.packs.find((p: any) => p.key === words)).toBeUndefined();
  });

  test('in a tic-tac-toe match a pack message reaches both players, a locked pack is refused, the third message locks', async () => {
    const room = await rpc(A, 'economy_game_friend_create', { key: 'stargame_tictactoe', room: 'bronze' });
    expect(room.status).toBe(200);
    matchId = room.data.matchId;
    a = await socket(A);
    b = await socket(B);
    await a.joinMatch(matchId);
    await b.joinMatch(matchId);
    await a.waitMatch(1);
    a.sendMatch(matchId, 6, { pack: words, item: 'gg', clientMessageId: 'm1' });
    const got = await b.waitMatch(7);
    expect(got).toMatchObject({ matchId, senderId: A.id, pack: words, packKind: 'text', item: 'gg', text: { fa: 'بازی خوبی بود' }, clientMessageId: 'm1' });
    expect((await a.waitMatch(7)).messageId).toBe(got.messageId);
    a.sendMatch(matchId, 6, { pack: faces, item: 'angry' });
    expect(await a.waitMatch(5, (d) => d.code === 'PACK_NOT_OWNED')).toMatchObject({ pack: faces });
    await eco.post('/users/' + A.id + '/items', { itemId: item.id });
    a.sendMatch(matchId, 6, { pack: faces, item: 'angry', clientMessageId: 'm2' });
    expect(await b.waitMatch(7, (d) => d.clientMessageId === 'm2')).toMatchObject({ packKind: 'emoji', assetId: 'emoji.angry' });
    a.sendMatch(matchId, 6, { pack: words, item: 'gg', clientMessageId: 'm3' });
    const lock = await a.waitMatch(5, (d) => d.code === 'SPAM_LOCKED');
    expect(lock.retryAfterMs).toBeGreaterThan(25_000);
    expect(b.matchData(7).filter((d) => d.clientMessageId === 'm3')).toHaveLength(0);
  });

  test('staff stop one player and close the match chat from the channels tab', async ({ page }) => {
    await open(page, '#/moderation/stickers/channels');
    await search(page, matchId);
    const row = root(page).locator(`[data-cm-channel="${matchId}"]`).locator('xpath=ancestor::tr');
    await expect(row.locator('[data-cm-status="open"]')).toContainText('باز');
    await expect(row.locator(`[data-cm-member="${A.id}"]`)).toContainText('۲ پیام');
    await row.locator(`[data-cm-can-send="${B.id}"]`).click();
    await expect(row.locator(`[data-cm-member="${B.id}"]`)).toContainText('فقط خواندن');
    b.sendMatch(matchId, 6, { pack: words, item: 'gg' });
    expect(await b.waitMatch(5, (d) => d.code === 'SEND_NOT_ALLOWED')).toBeTruthy();
    await row.locator(`[data-cm-close="${matchId}"]`).click();
    await page.locator('[data-economy="dialog"] button', { hasText: 'بستن گفتگو' }).click();
    await expect(toasts(page).last()).toContainText('گفتگو بسته شد');
    const ch = await base.get('/comms/channels/' + encodeURIComponent(matchId));
    expect(ch).toMatchObject({ status: 'closed', closeReason: 'staff' });
    a.sendMatch(matchId, 6, { pack: words, item: 'gg', clientMessageId: 'm9' });
    expect(await a.waitMatch(5, (d) => d.code === 'CHANNEL_CLOSED' || d.code === 'SPAM_LOCKED')).toBeTruthy();
  });
});
