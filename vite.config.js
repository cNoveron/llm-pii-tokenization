import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: 'static',
    emptyOutDir: true,
  },
  server: {
    port: 3000,
    proxy: {
      '/upload': 'http://localhost:8080',
      '/ask': 'http://localhost:8080',
      '/automate': 'http://localhost:8080',
      '/sample': 'http://localhost:8080',
    },
  },
});
