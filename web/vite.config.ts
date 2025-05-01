import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 8080, // Frontend port
    proxy: {
      '/api': {
        target: 'http://localhost:3000', // Backend port
        changeOrigin: true,
      },
    },
  },
  // Ensure React 19 JSX transform is used
  esbuild: {
    jsx: 'automatic',
  },
})
