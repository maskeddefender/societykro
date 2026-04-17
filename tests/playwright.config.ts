import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  timeout: 30000,
  retries: 0,
  workers: 1, // Serial execution — prevents OTP rate limit conflicts across test files
  reporter: [['list'], ['html', { outputFolder: 'test-results' }]],
  projects: [
    {
      name: 'api',
      testMatch: 'api/**/*.spec.ts',
    },
    {
      name: 'web-admin',
      testMatch: 'web-admin/**/*.spec.ts',
      use: {
        baseURL: 'http://localhost:3000',
        browserName: 'chromium',
        headless: true,
      },
    },
  ],
});
