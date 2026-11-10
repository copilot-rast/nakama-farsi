import { defineConfig } from '@playwright/test';
import { CONSOLE_URL } from './lib/env';

export default defineConfig({
  testDir: './specs',
  outputDir: './test-results',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  timeout: 120_000,
  expect: { timeout: 12_000 },
  reporter: [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],
  use: {
    baseURL: CONSOLE_URL,
    actionTimeout: 15_000,
    navigationTimeout: 30_000,
  },
  projects: [
    { name: 'setup', testMatch: /global\.setup\.ts/, teardown: 'teardown', use: { pageOpts: { signedIn: false } } as any },
    { name: 'teardown', testMatch: /global\.teardown\.ts/ },
    { name: 'console', testDir: './specs/console', dependencies: ['setup'], use: { pageOpts: { signedIn: true } } as any },
    { name: 'game', testDir: './specs/game', dependencies: ['setup'], use: { pageOpts: { signedIn: false } } as any },
  ],
});
