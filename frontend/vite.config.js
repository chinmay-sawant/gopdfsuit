import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

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