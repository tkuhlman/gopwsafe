import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  base: '/gopwsafe/',
  server: {
    host: true,
    port: 5173,
    strictPort: true,
  }
})
