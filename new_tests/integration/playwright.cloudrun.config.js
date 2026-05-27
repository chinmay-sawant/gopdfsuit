import { defineConfig } from '@playwright/test'

// Auth e2e: a fully self-hosted stack with authentication ENABLED end to end.
//   - auth-ms        (:9090) issues JWTs from email/password (SQLite in-memory)
//   - gopdfsuit back (:8080) enforces those JWTs (K_SERVICE + shared secret)
//   - frontend       (:3000) renders the login form (VITE_IS_CLOUD_RUN=true)
// Everything is local and deterministic, so the login flow runs for real — no
// external provider, no skips. merge.fail needs auth OFF and stays in
// playwright.config.js.
const AUTH_JWT_SECRET = 'e2e-shared-secret'

export default defineConfig({
  testDir: '.',
  testMatch: ['**/auth.spec.js'],
  timeout: 120_000,
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
      command: 'go run ./auth-ms',
      cwd: '../..',
      // Wait until /health responds (port-listen isn't enough on cold start).
      url: 'http://localhost:9090/health',
      reuseExistingServer: false,
      timeout: 180_000,
      env: {
        GOTOOLCHAIN: 'auto',
        AUTH_PORT: '9090',
        AUTH_DB_PATH: ':memory:',
        AUTH_JWT_SECRET,
        AUTH_CORS_ORIGIN: '*',
      },
    },
    {
      command: 'go run ./cmd/gopdfsuit',
      cwd: '../..',
      url: 'http://localhost:8080/',
      reuseExistingServer: false,
      timeout: 180_000,
      env: {
        GOTOOLCHAIN: 'auto',
        K_SERVICE: 'e2e-test',
        AUTH_JWT_SECRET,
      },
    },
    {
      command: 'npm run dev',
      cwd: '../../frontend',
      port: 3000,
      reuseExistingServer: false,
      timeout: 60_000,
      env: {
        VITE_IS_CLOUD_RUN: 'true',
        VITE_ENVIRONMENT: 'cloudrun',
        VITE_API_URL: 'http://localhost:8080',
        VITE_AUTH_URL: 'http://localhost:9090',
      },
    },
  ],
})
