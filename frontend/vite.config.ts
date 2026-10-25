/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// В режиме разработки запросы /api проксируются на Go-сервер.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: { '/api': process.env.VITE_API_TARGET ?? 'http://localhost:8080' },
  },
  preview: {
    port: 4173,
    proxy: { '/api': process.env.VITE_API_TARGET ?? 'http://localhost:8080' },
  },
  build: {
    rollupOptions: {
      output: {
        // раздельные чанки ускоряют загрузку и кэширование
        manualChunks: {
          react: ['react', 'react-dom', 'react-router-dom'],
          mui: ['@mui/material', '@mui/icons-material', '@emotion/react', '@emotion/styled'],
          charts: ['@mui/x-charts'],
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/setupTests.ts',
  },
});
