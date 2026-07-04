/// <reference types="vitest/config" />
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // 開発サーバーはRPCパスをGoサーバー(heic-converter serve)へ転送する。
    // ブラウザからは同一オリジンに見えるため、開発時はVITE_API_URLも
    // serve側の--allowed-origins(CORS)も不要になる。
    proxy: {
      '/heic.v1.ConvertService': 'http://localhost:8080',
      '/grpc.health.v1.Health': 'http://localhost:8080',
    },
  },
  test: {
    // Testing Libraryの自動クリーンアップ(afterEach)を有効にするために必要
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/**'],
      // 生成コードとエントリポイントはカバレッジ対象外
      exclude: ['src/gen/**', 'src/main.tsx', 'src/test/**'],
    },
  },
})
