import { defineConfig } from '@playwright/test'

// Local mode: backend + frontend run with auth DISABLED (no K_SERVICE).
// Runs merge.fail.spec.js only. The auth spec asserts Cloud Run behavior
// (401s + login screen) and runs under playwright.cloudrun.config.js.
export default defineConfig({
  testDir: '.',
  testIgnore: ['**/auth.spec.js'],
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
      // `url:` waits for a real 2xx/3xx response, not just the port being open.
      // Avoids the cold-start flake where `go run` starts listening before it
      // can actually handle requests.
      url: 'http://localhost:8080/',
      reuseExistingServer: true,
      timeout: 180_000,
      env: { GOTOOLCHAIN: 'auto' },
    },
    {
      command: 'pnpm run dev',
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
