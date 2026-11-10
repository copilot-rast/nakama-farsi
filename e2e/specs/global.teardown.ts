import { expect, test } from '../lib/fixtures';
import fs from 'node:fs';
import path from 'node:path';
import { cleanupLeftovers, configHashes, dropBackup, leftovers, restoreConfig, savedHashes } from '../lib/db';
import { STATE_DIR } from '../lib/env';

test('specs left shared config and data as they found it', async () => {
  const before = savedHashes();
  const adjusted: string[] = JSON.parse(fs.readFileSync(path.join(STATE_DIR, 'adjusted.json'), 'utf8'));
  const after = configHashes();
  const changed = Object.keys(before).filter((k) => before[k] !== after[k] && !adjusted.includes(k));
  const left = leftovers();
  if (Object.keys(left).length) cleanupLeftovers();
  restoreConfig();
  const restored = configHashes();
  dropBackup();
  console.log('config hashes', JSON.stringify({ changedBySpecs: changed, leftovers: left, restoredEqual: JSON.stringify(restored) === JSON.stringify(before) }));
  expect(restored, 'restore from the backup gives the exact rows of the start').toEqual(before);
  expect(changed, 'config rows changed by specs and not restored').toEqual([]);
  expect(left, 'e2e rows left behind').toEqual({});
});
