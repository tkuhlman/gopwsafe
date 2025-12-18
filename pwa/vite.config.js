import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    svelte(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['vite.svg', 'wasm_exec.js', 'gopwsafe.wasm.gz'],
      manifest: {
        name: 'GoPasswordSafe',
        short_name: 'GoPWSafe',
        description: 'Password Safe PWA powered by Go and WASM',
        theme_color: '#ffffff',
        icons: [
          {
            src: 'vite.svg',
            sizes: '192x192',
            type: 'image/svg+xml'
          },
          {
            src: 'vite.svg',
            sizes: '512x512',
            type: 'image/svg+xml'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,wasm,gz}'],
        maximumFileSizeToCacheInBytes: 5000000,
      }
    })
  ],
  base: '/gopwsafe/',
  server: {
    host: true,
    port: 5173,
    strictPort: true,
  }
})
