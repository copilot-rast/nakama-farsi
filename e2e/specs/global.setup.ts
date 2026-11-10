import { expect, test as setup } from '../lib/fixtures';
import fs from 'node:fs';
import path from 'node:path';
import { base } from '../lib/api';
import { backupConfig, dropBackup, restoreConfig, sql } from '../lib/db';
import { AUTH_FILE, STATE_DIR } from '../lib/env';
import { loginAs, ADMIN } from '../lib/console';

setup('back up shared config and sign in once', async ({ page }) => {
  if (sql("select count(*) from pg_namespace where nspname = 'zz_e2e_backup'") !== '0') {
    restoreConfig();
    dropBackup();
  }
  backupConfig();
  const adjust: string[] = [];
  const st = await base.get('/setup');
  if (st.status === 'pending') {
    await base.put('/setup', { status: 'completed' });
    adjust.push('H-setup');
  }
  fs.writeFileSync(path.join(STATE_DIR, 'adjusted.json'), JSON.stringify(adjust));
  await loginAs(page, ADMIN);
  await expect(page.locator('[data-economy-nav="economy"]').first()).toBeAttached();
  await page.evaluate(() => { localStorage.setItem('nakama.console.locale', 'fa-IR'); localStorage.setItem('color-mode', 'light'); });
  await page.context().storageState({ path: AUTH_FILE });
});
