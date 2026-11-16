import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api(?=\/|$)/, ''),
        },
      },
    },
    test: {
      environment: 'jsdom',
      environmentOptions: {
        jsdom: { url: 'http://localhost:5173/' },
      },
      setupFiles: ['./src/test/setup.ts'],
      restoreMocks: true,
      coverage: {
        provider: 'v8',
        reportsDirectory: 'coverage',
        include: ['src/**/*.{ts,tsx}'],
        exclude: ['src/main.tsx', 'src/test/**'],
      },
    },
  }
})
