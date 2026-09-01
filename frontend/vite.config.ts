import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/auth': 'http://localhost:8080',
      '/candles': 'http://localhost:8080',
      '/experiments': 'http://localhost:8080',
      '/search': 'http://localhost:8080',
      '/sentiment': 'http://localhost:8080',
      '/strategies': 'http://localhost:8080',
      '/markets': 'http://localhost:8080',
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
})
