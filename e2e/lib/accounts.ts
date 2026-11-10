import fs from 'node:fs';
import { expect } from '@playwright/test';
import { deleteUser } from './api';
import { sql } from './db';
import { SERVER_LOG } from './env';

export const TEST_PHONE_PREFIX = '+98999';

export function testPhone() {
  const n = String(Math.floor(Math.random() * 1e7)).padStart(7, '0');
  return { local: '0999' + n, e164: TEST_PHONE_PREFIX + n };
}

export function resetOtpLimits() {
  sql(`DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:%'`);
}

function codes(phone: string): string[] {
  const out: string[] = [];
  for (const line of fs.readFileSync(SERVER_LOG, 'utf8').split('\n')) {
    if (!line.includes('OTP sandbox code') || !line.includes(phone + ' ')) continue;
    const m = /OTP sandbox code (\d+) for /.exec(line);
    if (m) out.push(m[1]);
  }
  return out;
}

export function codeCount(phone: string) {
  return codes(phone).length;
}

export async function waitCode(phone: string, before: number): Promise<string> {
  let found = '';
  await expect.poll(() => {
    const list = codes(phone);
    if (list.length > before) found = list[list.length - 1];
    return list.length;
  }, { message: 'sandbox code for ' + phone + ' in ' + SERVER_LOG, timeout: 10_000 }).toBeGreaterThan(before);
  return found;
}

export async function removePhones() {
  const ids = sql(`SELECT user_id FROM publication_phone_account WHERE phone LIKE '${TEST_PHONE_PREFIX}%'`).split('\n').filter(Boolean);
  for (const id of ids) await deleteUser(id);
  sql(`DELETE FROM publication_phone_account WHERE phone LIKE '${TEST_PHONE_PREFIX}%'`);
  sql(`DELETE FROM publication_otp_code WHERE phone LIKE '${TEST_PHONE_PREFIX}%'`);
  resetOtpLimits();
}

export function otpSandboxOn() {
  return sql(`SELECT is_enabled AND sandbox FROM publication_otp_settings WHERE id=1`) === 't';
}
