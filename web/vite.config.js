import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: path.resolve(process.env.HOME, '.makasero', 'web-frontend'),
    emptyOutDir: true,
    sourcemap: true,
  },
})
