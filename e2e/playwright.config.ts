import { defineConfig } from '@playwright/test';
import { CONSOLE_URL } from './lib/env';
import { sampleSpecIgnores } from './lib/samples';

const sampleIgnores = sampleSpecIgnores();

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
  webServer: {
    command: 'node fake-github/server.mjs',
    url: 'http://127.0.0.1:7489/__fake/scenario',
    reuseExistingServer: true,
    stdout: 'ignore',
    stderr: 'pipe',
    timeout: 15_000,
  },
  projects: [
    { name: 'setup', testMatch: /global\.setup\.ts/, teardown: 'teardown', use: { pageOpts: { signedIn: false } } as any },
    { name: 'teardown', testMatch: /global\.teardown\.ts/ },
    { name: 'console', testDir: './specs/console', testIgnore: sampleIgnores, dependencies: ['setup'], use: { pageOpts: { signedIn: true } } as any },
    { name: 'game', testDir: './specs/game', testIgnore: sampleIgnores, dependencies: ['setup'], use: { pageOpts: { signedIn: false } } as any },
  ],
});
