import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './specs',
  outputDir: '../test-results/cluster',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  timeout: 180_000,
  expect: { timeout: 15_000 },
  globalSetup: './setup.ts',
  globalTeardown: './teardown.ts',
  reporter: [['list']],
});
