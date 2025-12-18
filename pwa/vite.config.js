import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { VitePWA } from 'vite-plugin-pwa'
import { execSync } from 'child_process'

// Get version from git
let version = '0.0.0-dev'
try {
  version = execSync('git describe --tags --dirty --always').toString().trim()
} catch (e) {
  console.warn('Could not get git version:', e)
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    svelte(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['wasm_exec.js', 'gopwsafe.wasm.gz', 'lock-combined.svg', 'lock-foreground.svg', 'pwa-icon.png'],
      manifest: {
        name: 'GoPasswordSafe',
        short_name: 'GoPWSafe',
        description: 'Password Safe PWA powered by Go and WASM',
        theme_color: '#ffffff',
        display: 'standalone',
        background_color: '#ffffff',
        start_url: '/gopwsafe/',
        icons: [
          {
            src: 'lock-combined.svg',
            sizes: '512x512',
            type: 'image/svg+xml',
            purpose: 'maskable'
          },
          {
            src: 'lock-foreground.svg',
            sizes: '512x512',
            type: 'image/svg+xml',
            purpose: 'any'
          },
          {
            src: 'pwa-icon.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'pwa-icon.png',
            sizes: '512x512',
            type: 'image/png'
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
  },
  define: {
    __APP_VERSION__: JSON.stringify(version)
  }
})
