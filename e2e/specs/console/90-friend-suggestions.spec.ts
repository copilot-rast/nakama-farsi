import { expect, test } from '../../lib/fixtures';
import { deleteUser, devicePlayer, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { CONSOLE_URL } from '../../lib/env';
import { field, openSettings, root, toasts } from '../../lib/console';

const k = uid();
const users: string[] = [];
const matches: string[] = [];
let V: Player;
let X: Player;
let Y: Player;
let F: Player;
let M: Player;
let saved = '';

test.describe.configure({ mode: 'serial' });

function q(v: string) {
  return "'" + v.replace(/'/g, "''") + "'";
}

function finished(players: Player[], ago: number) {
  const id = 'e2e-' + k + '-' + matches.length + '.nf';
  matches.push(id);
  sql(`INSERT INTO economy_game_match (id, game_key, status, players, ended_at) VALUES (${q(id)}, ${q('e2e_' + k)}, 'finished', ${players.length}, now() - INTERVAL '${ago} seconds')`);
  players.forEach((p, i) => sql(`INSERT INTO economy_game_entry (match_id, user_id, status, created_at) VALUES (${q(id)}, '${p.id}', 'free', now() + INTERVAL '${i} milliseconds')`));
}

test.beforeAll(async () => {
  saved = sql(`SELECT COALESCE((SELECT value::text FROM base_friend_suggest_settings WHERE id=1), '')`);
  const mk = async (tag: string) => {
    const p = await devicePlayer('e2e-fs-' + tag + '-' + k, ('fs' + tag + k).replace(/_/g, '').slice(0, 20));
    users.push(p.id);
    return p;
  };
  V = await mk('v');
  X = await mk('x');
  Y = await mk('y');
  F = await mk('f');
  M = await mk('m');
  await updateAccount(V, { lang_tag: 'fa' });
  for (const [a, b] of [[V, F], [F, M]]) {
    sql(`INSERT INTO user_edge (source_id, destination_id, state, position, update_time) VALUES ('${a.id}','${b.id}',0,${Date.now()}1,now()), ('${b.id}','${a.id}',0,${Date.now()}2,now())`);
  }
  finished([V, X, Y], 40);
  finished([V, F], 30);
});

test.afterAll(async () => {
  if (saved) sql(`UPDATE base_friend_suggest_settings SET value=${q(saved)}::jsonb WHERE id=1`);
  else sql(`DELETE FROM base_friend_suggest_settings`);
  for (const id of matches) sql(`DELETE FROM economy_game_match WHERE id=${q(id)}`);
  for (const id of users) await deleteUser(id);
});

test.describe('friend suggestions', () => {
  test('finished matches become co-play encounters and co-players lead the list, friends are left out', async () => {
    await expect.poll(() => Number(sql(`SELECT count(*) FROM base_coplay WHERE user_id='${V.id}'`)), { timeout: 60_000, intervals: [2000] }).toBe(3);
    const res = await rpc(V, 'base_friend_suggestions', {});
    expect(res.status).toBe(200);
    expect(res.data.enabled).toBe(true);
    const list = res.data.suggestions as any[];
    const ids = list.map((s) => s.userId);
    expect(ids.slice(0, 2).sort()).toEqual([X.id, Y.id].sort());
    expect(list[0]).toMatchObject({ source: 'coplay' });
    expect(list[0].code).toMatch(/^[A-Z0-9]{6}$/);
    expect(list.find((s) => s.userId === M.id)).toMatchObject({ source: 'mutual', mutual: 1 });
    expect(ids).not.toContain(F.id);
    expect(ids).not.toContain(V.id);
    expect(sql(`SELECT shows FROM base_friend_suggest WHERE user_id='${V.id}' AND candidate_id='${X.id}'`)).toBe('1');
    await rpc(V, 'base_friend_suggestions', {});
    expect(sql(`SELECT shows FROM base_friend_suggest WHERE user_id='${V.id}' AND candidate_id='${X.id}'`)).toBe('1');
  });

  test('the friends tab saves list sizes and the player card rebuilds the list with a reserve', async ({ page }) => {
    await openSettings(page, 'friends');
    const form = page.locator('[data-settings-form="friend-suggestions"]');
    await expect(form).toContainText('پیشنهاد دوستی');
    const rules = page.locator('[data-settings-form="friends"]');
    await expect(rules).toContainText('پیام خصوصی فقط بین دوستان');
    await expect(rules).toContainText('وضعیت آنلاین فقط برای دوستان');
    await field(form, 'اندازهٔ فهرست').locator('input').fill('2');
    await field(form, 'ذخیره').locator('input').fill('1');
    await form.locator('[data-settings-save]').click();
    await expect(toasts(page).last()).toBeVisible();
    await expect.poll(() => sql(`SELECT value->>'visible' || '/' || (value->>'reserve') FROM base_friend_suggest_settings`)).toBe('2/1');

    await page.goto(CONSOLE_URL + '/#/players/' + V.id);
    const card = page.locator('[data-player-social]');
    await expect(card).toBeVisible({ timeout: 20_000 });
    await expect(card.locator('[data-ps-coplay]')).toHaveAttribute('data-ps-coplay', '3');
    await expect(card).toContainText('هم‌بازی');
    await card.locator('[data-ps-action="regenerate"]').click();
    await expect(card.locator('[data-ps-visible]')).toHaveAttribute('data-ps-visible', '2', { timeout: 20_000 });
    await expect(card.locator('[data-ps-reserve]')).toHaveAttribute('data-ps-reserve', '1');
    expect(sql(`SELECT generated_by <> '' FROM base_friend_suggest_list WHERE user_id='${V.id}'`)).toBe('t');

    const short = await rpc(V, 'base_friend_suggestions', {});
    expect(short.data.suggestions).toHaveLength(2);
  });
});
