import { expect, type Page } from '@playwright/test';
import { consoleApi, must, rpc, type Player } from './api';
import { CONSOLE_URL } from './env';

export const PNG = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAIAAABLbSncAAAAEUlEQVR4nGP4f9IGK2IYWhIAmHaBAfyC0NEAAAAASUVORK5CYII=', 'base64');

export async function uploadAs(player: Player, data: Buffer, name: string) {
  const begin = await rpc(player, 'base_support_upload_begin', { name, size: data.length });
  expect(begin.status, JSON.stringify(begin.data)).toBe(200);
  const size = begin.data.chunkBytes as number;
  for (let i = 0; i < begin.data.parts; i++) {
    const part = await rpc(player, 'base_support_upload_part', { uploadId: begin.data.uploadId, index: i, data: data.subarray(i * size, (i + 1) * size).toString('base64') });
    expect(part.status).toBe(200);
  }
  const done = await rpc(player, 'base_support_upload_finish', { uploadId: begin.data.uploadId });
  return done;
}

export type ConsoleUser = { user: string; pass: string };

export async function makeConsoleUser(name: string, acl: Record<string, { read: boolean; write: boolean }>): Promise<ConsoleUser> {
  const user = { user: name, pass: 'E2eQa-' + Math.random().toString(36).slice(2, 10) + '9!' };
  const add = await consoleApi('POST', '/v2/console/user', { username: user.user, email: user.user + '@example.com', acl });
  expect(add.status).toBeLessThan(300);
  const reset = await must(consoleApi('POST', '/v2/console/user/' + user.user + '/reset/password'));
  const pc = await fetch(CONSOLE_URL + '/v2/console/authenticate/password_change', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: reset.code || add.data?.token, password: user.pass }) });
  expect(pc.ok).toBe(true);
  return user;
}

export async function dropConsoleUser(name: string) {
  await consoleApi('DELETE', '/v2/console/user?username=' + name);
}

export async function signIn(p: Page, user: ConsoleUser, waitNav = 'support') {
  await p.goto(CONSOLE_URL + '/#/login');
  await p.locator('input[name="username"]').fill(user.user);
  await p.locator('input[name="password"]').fill(user.pass);
  await p.keyboard.press('Enter');
  await p.locator(`.x-drawer [data-economy-nav="${waitNav}"]`).first().waitFor({ state: 'attached', timeout: 20_000 });
}

export async function bearerOf(p: Page) {
  const tok = await p.evaluate(() => JSON.parse(localStorage.getItem('SESSION_TOKENS') || '{}'));
  return 'Bearer ' + tok.accessToken;
}
