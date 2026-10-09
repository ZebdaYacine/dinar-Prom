import {defineConfig, devices} from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: [['list']],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://127.0.0.1:8180',
    locale: 'ar-DZ',
    ...devices['Desktop Chrome'],
    launchOptions: {executablePath: process.env.CHROME_BIN ?? '/usr/bin/google-chrome'}
  }
});
