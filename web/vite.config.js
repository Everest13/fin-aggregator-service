import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/banks': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/upload-csv': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/users': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/categories': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/transactions': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/transaction-types': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/monzo/transactions': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/monzo/auth-url': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/transactions/{transaction_id}': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/monzo/account': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/monzo/callback': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      }
    }
  }
})