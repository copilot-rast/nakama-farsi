import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, eco, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { Chat, notice } from '../../lib/chat';
import { sql } from '../../lib/db';

const k = uid();
const made = { users: [] as string[], packs: [] as string[], items: [] as string[], types: [] as string[] };
const sockets: Chat[] = [];
let savedSticker = '';
let savedQuota = '';
let free = '';
let vip = '';
let item: any;
let A: Player;
let B: Player;

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

async function socket(p: Player) {
  const c = await Chat.open(p);
  sockets.push(c);
  return c;
}

async function stickerSettings(patch: Record<string, unknown>) {
  return base.put('/stickers/settings', { enabled: true, direct: true, groups: true, rooms: true, countQuota: true, ...patch });
}

test.beforeAll(async () => {
  savedSticker = sql(`SELECT settings::text FROM base_sticker_settings WHERE id=1`);
  savedQuota = sql(`SELECT settings::text || '|' || updated_at::text FROM base_chat_quota_settings WHERE id=1`);
  const type = await base.post('/categories', { key: 'e2e_' + k + '_stt', keyType: 'item_type', name: { fa: 'بستهٔ استیکر' }, description: {}, sort: 0, metadata: {} });
  made.types.push(type.id);
  item = await eco.post('/items', { key: 'e2e_' + k + '_stickers', typeId: type.id, name: { fa: 'استیکرهای ویژه', en: 'Special stickers' } });
  made.items.push(item.id);
  free = 'e2e_' + k + '_free';
  vip = 'e2e_' + k + '_vip';
  const a = await base.post('/stickers/packs', { key: free, name: { fa: 'شاد', en: 'Happy' }, sort: -1000, stickers: [{ key: 'smile', name: { fa: 'لبخند' }, imageUrl: '/e2e/smile.png' }, { key: 'laugh', imageUrl: '/e2e/laugh.png' }] });
  const b = await base.post('/stickers/packs', { key: vip, name: { fa: 'ویژه', en: 'Special' }, sort: -999, access: 'item', itemKey: item.key, stickers: [{ key: 'crown', imageUrl: '/e2e/crown.png' }] });
  made.packs.push(a.id, b.id);
  A = await devicePlayer('e2e-st-a-' + k, ('sta' + k).replace(/_/g, '').slice(0, 20));
  B = await devicePlayer('e2e-st-b-' + k, ('stb' + k).replace(/_/g, '').slice(0, 20));
  made.users.push(A.id, B.id);
  await updateAccount(A, { lang_tag: 'fa' });
  await updateAccount(B, { lang_tag: 'en' });
  await stickerSettings({});
  await base.put('/chat-quota/settings', { enabled: false });
});

test.afterAll(async () => {
  for (const s of sockets) s.close();
  if (savedSticker) sql(`UPDATE base_sticker_settings SET settings=${q(savedSticker)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_sticker_settings`);
  if (savedQuota) {
    const i = savedQuota.lastIndexOf('|');
    sql('UPDATE base_chat_quota_settings SET settings=' + q(savedQuota.slice(0, i)) + '::jsonb, updated_at=' + q(savedQuota.slice(i + 1)) + '::timestamptz WHERE id=1');
  }
  for (const id of made.packs) await base.del('/stickers/packs/' + id);
  for (const id of made.users) await deleteUser(id);
  if (made.items.length) sql(`DELETE FROM economy_user_item WHERE item_id IN (${made.items.map((x) => `'${x}'`).join(',')})`);
  for (const id of made.items) await eco.del('/items/' + id);
  for (const id of made.types) await base.del('/categories/' + id);
});

test.describe('game: stickers over the socket and RPC', () => {
  test('base_stickers lists the packs with the free one open and the item one locked with its item', async () => {
    const res = await rpc(A, 'base_stickers', { lang: 'fa' });
    expect(res.status).toBe(200);
    expect(res.data).toMatchObject({ enabled: true, chats: { direct: true, groups: true, rooms: true }, countQuota: true });
    const packs = (res.data.packs as any[]).filter((p) => p.key === free || p.key === vip);
    expect(packs.map((p) => p.key)).toEqual([free, vip]);
    expect(packs[0]).toMatchObject({ name: 'شاد', access: 'free', owned: true });
    expect(packs[0].stickers.map((s: any) => s.ref)).toEqual([free + '/smile', free + '/laugh']);
    expect(packs[0].stickers[0]).toMatchObject({ name: 'لبخند', imageUrl: '/e2e/smile.png' });
    expect(packs[1]).toMatchObject({ name: 'ویژه', access: 'item', owned: false, itemKey: item.key, item: { key: item.key, name: 'استیکرهای ویژه' } });
    const en = await rpc(B, 'base_stickers', {});
    expect((en.data.packs as any[]).find((p) => p.key === free).name).toBe('Happy');
  });

  test('a free sticker is sent with its picture, a locked one is refused until the item is owned, text is untouched', async () => {
    const a = await socket(A);
    const dm = await a.joinDirect(B.id);
    const sent = await a.sendContent({ sticker: free + '/smile' });
    expect(sent.ok).toBe(true);
    const stored = JSON.parse(sql(`SELECT content FROM message WHERE id='${(sent as any).id}'`));
    expect(stored).toEqual({ sticker: free + '/smile', stickerPack: free, stickerUrl: '/e2e/smile.png' });

    const locked = await a.sendContent({ sticker: vip + '/crown' });
    expect(locked.ok).toBe(false);
    const n = notice((locked as any).message);
    expect(n).toMatchObject({ code: 'STICKER_LOCKED', lang: 'fa', meta: { pack: vip, itemKey: item.key } });
    expect(n.message).toContain('استیکر');
    expect(notice(((await a.sendContent({ sticker: free + '/none' })) as any).message).code).toBe('STICKER_NOT_FOUND');

    await eco.post('/users/' + A.id + '/items', { itemId: item.id });
    expect((await a.sendContent({ sticker: vip + '/crown' })).ok).toBe(true);
    expect((await rpc(A, 'base_stickers', {})).data.packs.find((p: any) => p.key === vip).owned).toBe(true);

    const text = await a.send('سلام ' + k);
    expect(text.ok).toBe(true);
    expect(JSON.parse(sql(`SELECT content FROM message WHERE id='${(text as any).id}'`))).toEqual({ text: 'سلام ' + k });
    expect(dm.startsWith('4.')).toBe(true);
  });

  test('settings close private chats to stickers and keep rooms open; turning stickers off refuses them everywhere', async () => {
    await stickerSettings({ direct: false });
    const b = await socket(B);
    await b.joinDirect(A.id);
    const n = notice(((await b.sendContent({ sticker: free + '/smile' })) as any).message);
    expect(n).toMatchObject({ code: 'STICKER_NOT_ALLOWED_HERE', lang: 'en', meta: { chat: 'dm' } });
    expect(n.message).toContain('not allowed');
    const room = await socket(B);
    await room.join('e2e-st-' + k.replace(/_/g, '-').toLowerCase());
    expect((await room.sendContent({ sticker: free + '/laugh' })).ok).toBe(true);

    await stickerSettings({ enabled: false });
    expect(notice(((await room.sendContent({ sticker: free + '/laugh' })) as any).message).code).toBe('STICKERS_DISABLED');
    expect((await rpc(B, 'base_stickers', {})).data).toMatchObject({ enabled: false, packs: [] });
    await stickerSettings({});
  });

  test('stickers outside the quota when countQuota is off, inside when on; private chat not counted when countDirect is off', async () => {
    await base.put('/chat-quota/settings', { enabled: true, dailyMessages: 1, countMode: 'messages' });
    const a = await socket(A);
    await a.joinDirect(B.id);
    expect((await a.send('یک')).ok).toBe(true);
    const over = await a.send('دو');
    expect(over.ok).toBe(false);
    expect(notice((over as any).message).code).toBe('DM_QUOTA_EXCEEDED');

    await stickerSettings({ countQuota: false });
    expect((await a.sendContent({ sticker: free + '/smile' })).ok).toBe(true);
    await stickerSettings({ countQuota: true });
    expect(notice(((await a.sendContent({ sticker: free + '/smile' })) as any).message).code).toBe('DM_QUOTA_EXCEEDED');

    await base.put('/chat-quota/settings', { enabled: true, dailyMessages: 1, countMode: 'messages', countDirect: false });
    expect((await a.send('سه')).ok).toBe(true);
    const me = await rpc(A, 'base_chat_quota', {});
    expect(me.data.counts).toMatchObject({ direct: false });
  });
});
