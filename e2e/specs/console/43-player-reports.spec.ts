import { expect, test } from '../../lib/fixtures';
import { base, deleteUser, devicePlayer, rpc, uid, updateAccount, type Player } from '../../lib/api';
import { sql } from '../../lib/db';
import { field, open, openSettings, root, rows, settle, toasts } from '../../lib/console';

const k = uid();
const users: string[] = [];
const titles: string[] = [];
let title: any;
let A: Player;
let B: Player;
let C: Player;
let reportId = '';
let otherId = '';

test.describe.configure({ mode: 'serial' });

async function player(name: string): Promise<Player> {
  const p = await devicePlayer('e2e-pr-' + name + '-' + k, ('pr' + name + k).replace(/_/g, '').slice(0, 30));
  users.push(p.id);
  await updateAccount(p, { lang_tag: 'fa', display_name: 'بازیکن ' + name + ' ' + k.slice(-4) });
  return p;
}

test.beforeAll(async () => {
  title = await base.post('/report-titles', { key: 'e2e_' + k + '_cheat', name: { fa: 'تقلب ' + k.slice(-4), en: 'Cheating ' + k.slice(-4) }, description: {}, isActive: true, sort: 1 });
  titles.push(title.id);
  A = await player('a');
  B = await player('b');
  C = await player('c');
});

test.afterAll(async () => {
  sql(`DELETE FROM base_player_report WHERE title_key LIKE 'e2e_%' OR reporter_id IN (${users.map((u) => `'${u}'`).join(',')}); DELETE FROM cluster_rate_limit WHERE key LIKE 'player-report:%';`);
  for (const id of users) await deleteUser(id);
  for (const id of titles) await base.del('/report-titles/' + id);
});

test.describe('player reports: client RPCs', () => {
  test('report a player with context, refusals, duplicate window and my reports', async () => {
    const self = await rpc(A, 'base_report_player', { userId: A.id, titleKey: title.key });
    expect(self.status).toBe(400);
    expect(JSON.stringify(self.data)).toContain('CANNOT_REPORT_SELF');
    const badTitle = await rpc(A, 'base_report_player', { userId: B.id, titleKey: 'nope_' + k });
    expect(JSON.stringify(badTitle.data)).toContain('REPORT_TITLE_NOT_FOUND');
    const long = await rpc(A, 'base_report_player', { userId: B.id, titleKey: title.key, note: 'ب'.repeat(1001) });
    expect(JSON.stringify(long.data)).toContain('TEXT_TOO_LONG');

    const note = 'در مسابقه از ربات استفاده کرد. ' + 'ب'.repeat(900);
    const first = await rpc(A, 'base_report_player', { userId: B.id, titleKey: title.key, note, matchId: 'e2e-match-' + k, contentId: 'replay-7', gameKey: 'stargame' });
    expect(first.status).toBe(200);
    expect(first.data.duplicate).toBe(false);
    reportId = first.data.reportId;
    const dup = await rpc(A, 'base_report_player', { userId: B.id, titleKey: title.key });
    expect(dup.data).toEqual({ reportId, duplicate: true });
    const second = await rpc(C, 'base_report_player', { userId: B.id, titleKey: title.key, gameKey: 'tictactoe' });
    otherId = second.data.reportId;
    expect(second.data.duplicate).toBe(false);

    const mine = (await rpc(A, 'base_my_reports', {})).data;
    expect(mine.total).toBe(1);
    expect(mine.items[0]).toMatchObject({ id: reportId, userId: B.id, status: 'open', gameKey: 'stargame', matchId: 'e2e-match-' + k, title: 'تقلب ' + k.slice(-4) });
    expect(mine.items[0].action).toBeUndefined();
    expect((await rpc(B, 'base_my_reports', {})).data.total).toBe(0);
  });
});

test.describe('player reports: console review', () => {
  test('list with context chips and badge, take for review, internal note, mute with closing the other report, audit log and the reporter sees the outcome', async ({ page }) => {
    await open(page, '#/moderation/player-reports');
    await settle(page);
    await expect(root(page).locator('h1')).toHaveText('گزارش بازیکن‌ها');
    const nav = page.locator('.pl-side a[href^="#/moderation/player-reports"]').first();
    await expect(nav).toBeVisible();
    const row = rows(page).filter({ hasText: 'بازیکن b ' + k.slice(-4) }).filter({ hasText: 'بازی: stargame' });
    await expect(row).toHaveCount(1);
    await expect(row).toContainText('تقلب ' + k.slice(-4));
    await expect(row).toContainText('باز');
    await row.click();
    await expect(page).toHaveURL(new RegExp('#/moderation/player-reports/' + reportId));
    const detail = root(page).locator(`[data-player-report="${reportId}"]`);
    await expect(detail.locator('[data-pr-note]')).toContainText('در مسابقه از ربات استفاده کرد');
    await expect(detail).toContainText('مسابقه: e2e-match-');
    await expect(detail).toContainText('محتوا: replay-7');
    await expect(detail.locator('[data-pr-counts]')).toContainText('۲ گزارش بازیکن');
    await expect(detail).toContainText('گزارش‌های دیگر علیه او');

    await detail.locator('[data-pr-review="open"]').click();
    await expect(toasts(page).last()).toContainText('در حال بررسی شما');
    await expect(detail).toContainText('در حال بررسی');
    await expect(detail.locator('[data-pr-review="reviewing"]')).toBeVisible();

    await detail.locator('[data-pr-note-add]').click();
    await expect(detail).toContainText('متنی بنویسید');
    await detail.getByRole('textbox', { name: 'افزودن یادداشت' }).fill('ویدیوی بازی را دیدم');
    await detail.locator('[data-pr-note-add]').click();
    await expect(detail.locator('[data-pr-notes]')).toContainText('ویدیوی بازی را دیدم');

    await detail.locator('[data-pr-actions] button[data-value="mute"]').click();
    const form = detail.locator('[data-restrict-form="mute"]');
    await expect(form).toBeVisible();
    await form.locator('[data-duration] button', { hasText: '۱ ساعت' }).click();
    await detail.getByText('گزارش‌های باز دیگر علیه این بازیکن').click();
    await detail.locator('[data-pr-submit]').click();
    await expect(toasts(page).last()).toContainText('بازیکن بی‌صدا شد');
    await expect(detail).toContainText('رسیدگی شد');

    const player = await base.get('/players/' + B.id);
    expect(player.mute).not.toBeNull();
    expect(player.playerReports).toMatchObject({ against: 2, open: 0 });
    const other = await base.get('/player-reports/' + otherId);
    expect(other).toMatchObject({ status: 'resolved', action: 'mute' });
    expect((await rpc(A, 'base_my_reports', {})).data.items[0].status).toBe('resolved');
    const audit = Number(sql(`SELECT count(*) FROM console_audit_log WHERE message LIKE '%player-reports%' AND create_time > now() - interval '5 minutes'`));
    expect(audit).toBeGreaterThanOrEqual(3);
    expect(sql(`SELECT report_id::text FROM base_restriction WHERE user_id='${B.id}' AND kind='mute' ORDER BY created_at DESC LIMIT 1`)).toBe(reportId);

    const again = await base.post('/player-reports/' + reportId + '/resolve', { action: 'dismiss' }).catch((e) => String(e));
    expect(String(again)).toContain('REPORT_CLOSED');
  });

  test('player card links to player reports and the status filter', async ({ page }) => {
    await open(page, '#/moderation/player-reports?status=resolved&targetId=' + B.id);
    await settle(page);
    await expect(rows(page)).toHaveCount(2);
    await page.goto(page.url().split('#')[0] + '#/players/' + B.id);
    const link = page.locator('[data-player-mod] [data-player-reports-link]');
    await expect(link).toContainText('۲ گزارش بازیکن علیه او');
  });
});

test.describe('player reports: rules', () => {
  test('report rules section: per hour, note required, trusted tag; limits in base_my_reports', async ({ page }) => {
    const before = await base.get('/report-rules');
    try {
      await openSettings(page, 'report-rules');
      const form = page.locator('[data-settings-form="report-rules"]');
      await expect(form.locator('form')).toBeVisible();
      await expect(field(form, 'گزارش در ساعت').locator('input')).toHaveValue(/^(10|۱۰)$/);
      await field(form, 'گزارش در ساعت').locator('input').fill('1');
      await form.getByText('توضیح بازیکن لازم است').click();
      await form.locator('[data-settings-save]').click();
      await expect(toasts(page).last()).toContainText('ذخیره');
      expect(await base.get('/report-rules')).toMatchObject({ perHour: 1, noteRequired: true });
      sql(`DELETE FROM cluster_rate_limit WHERE key LIKE 'player-report:%'`);
      const R = await player('r');
      const noNote = await rpc(R, 'base_report_player', { userId: C.id, titleKey: title.key });
      expect(JSON.stringify(noNote.data)).toContain('NOTE_REQUIRED');
      expect((await rpc(R, 'base_report_player', { userId: C.id, titleKey: title.key, note: 'تقلب' })).status).toBe(200);
      const over = await rpc(R, 'base_report_player', { userId: A.id, titleKey: title.key, note: 'باز هم' });
      expect(over.status).toBe(429);
      expect((await rpc(R, 'base_my_reports', {})).data.limits).toMatchObject({ perHour: 1, remaining: 0, noteRequired: true, trusted: false });
    } finally {
      await base.put('/report-rules', { perHour: before.perHour, duplicateHours: before.duplicateHours, minAccountHours: before.minAccountHours, noteRequired: before.noteRequired, trustedTagIds: before.trustedTagIds || [], trustedPerHour: before.trustedPerHour });
      sql(`UPDATE base_report_rules SET updated_by='${String(before.updatedBy || '').replace(/'/g, "''")}', updated_at='${before.updatedAt}' WHERE id=1`);
    }
  });
});
