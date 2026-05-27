import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath } from 'url'
import path from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  // Set the third parameter to '' to load all env regardless of the `VITE_` prefix.
  const env = loadEnv(mode, '../', '')
  const target = env.VITE_API_URL || 'http://localhost:8080'

  return {
    plugins: [react()],
    base: '/gopdfsuit/',
    envDir: '../',
    resolve: {
      alias: {
        '@testing-library/jest-dom': path.resolve(__dirname, 'node_modules/@testing-library/jest-dom'),
        '@testing-library/react': path.resolve(__dirname, 'node_modules/@testing-library/react'),
        '@testing-library/user-event': path.resolve(__dirname, 'node_modules/@testing-library/user-event'),
      },
    },
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: path.resolve(__dirname, '../new_tests/frontend/setup.js'),
    },
    build: {
      outDir: '../docs',
      emptyOutDir: true,
      rollupOptions: {
        output: {
          manualChunks: undefined,
        },
      },
    },
    server: {
      port: 3000,
      fs: {
        allow: ['..']
      },
      proxy: {
        '/api': {
          target: target,
          changeOrigin: true,
        },
        '/static': {
          target: target,
          changeOrigin: true,
        },
      },
    },
  }
})