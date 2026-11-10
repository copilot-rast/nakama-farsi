import { expect, type Page } from '@playwright/test';
import { GAME_URL } from './env';

export type GamePlayer = { id: string; username: string; deviceId: string; token: string };

export async function openGame(page: Page, deviceId: string, opts: { lang?: string; query?: string } = {}): Promise<GamePlayer> {
  const origin = new URL(GAME_URL).origin;
  await page.addInitScript(([dev, lang, org]) => {
    if (location.origin !== org) return;
    if (sessionStorage.getItem('e2e-game')) return;
    localStorage.setItem('stargame.device', dev);
    if (lang) localStorage.setItem('stargame.lang', lang);
    sessionStorage.setItem('e2e-game', '1');
  }, [deviceId, opts.lang || '', origin]);
  await page.goto(GAME_URL + (opts.query || ''));
  if (!opts.query) await waitHome(page);
  return player(page);
}

export async function waitHome(page: Page) {
  await expect(page.locator('#screen-home')).toBeVisible({ timeout: 20_000 });
  await expect(page.locator('#balance')).not.toHaveText('…');
}

export async function player(page: Page): Promise<GamePlayer> {
  await page.waitForFunction(() => (window as any).eval('typeof state !== "undefined" && !!state.token'));
  const s = await page.evaluate(() => (window as any).eval('({ token: state.token, username: state.username, device: localStorage.getItem("stargame.device") })'));
  const claims = JSON.parse(Buffer.from(s.token.split('.')[1], 'base64url').toString('utf8'));
  return { id: claims.uid, username: s.username, deviceId: s.device, token: s.token };
}

export async function gameRpc(page: Page, id: string, payload: unknown = {}) {
  return page.evaluate(async ([id, payload]) => {
    const w = window as any;
    try {
      return { ok: true, data: await w.eval('rpc')(id, payload, false) };
    } catch (e: any) {
      return { ok: false, code: e.code, status: e.status };
    }
  }, [id, payload] as [string, unknown]);
}

export function toast(page: Page) {
  return page.locator('#toast');
}

export async function screen(page: Page, name: string) {
  await expect(page.locator('#screen-' + name)).toBeVisible();
}

export function persian(n: number | string) {
  return String(n).replace(/\d/g, (d) => '۰۱۲۳۴۵۶۷۸۹'[Number(d)]);
}

export async function balance(page: Page): Promise<number> {
  const t = await page.locator('#balance').innerText();
  return Number(t.replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 0x06f0)).replace(/[^\d-]/g, ''));
}
