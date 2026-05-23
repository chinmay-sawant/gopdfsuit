import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: 'http://localhost:3000',
    headless: true,
    trace: 'on-first-retry',
    video: 'retain-on-failure',
  },
  webServer: [
    {
      command: 'go run ./cmd/gopdfsuit',
      cwd: '../..',
      port: 8080,
      reuseExistingServer: true,
      timeout: 180_000,
      env: { GOTOOLCHAIN: 'go1.24.11+auto' },
    },
    {
      command: 'npm run dev',
      cwd: '../../frontend',
      port: 3000,
      reuseExistingServer: true,
      timeout: 60_000,
      env: {
        VITE_IS_CLOUD_RUN: 'false',
        VITE_ENVIRONMENT: 'local',
        VITE_API_URL: 'http://localhost:8080',
      },
    },
  ],
})
