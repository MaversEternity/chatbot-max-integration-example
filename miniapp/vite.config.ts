import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  base: process.env.GITHUB_PAGES ? '/chatbot-max-integration-example/' : '/miniapp/',
  build: {
    outDir: process.env.GITHUB_PAGES ? 'dist' : '../web/miniapp',
    emptyOutDir: true,
  },
})
