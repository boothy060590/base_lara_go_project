import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import fs from 'fs';
import path from 'path';

// ESM-compatible __dirname
import { fileURLToPath } from 'url';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export default defineConfig(({ mode }) => {
  // Load env variables for the current mode
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [vue()],
    css: {
      preprocessorOptions: {
        scss: {
          // Modern SASS API
          api: 'modern-compiler',
        },
      },
    },
    server: {
      port: env.VITE_VIRTUAL_PORT ? Number(env.VITE_VIRTUAL_PORT) : 5173,
      host: env.VITE_VIRTUAL_HOST || '0.0.0.0',
      https: {
        key: fs.readFileSync('/app/certs/app.baselaragoproject.test.key'),
        cert: fs.readFileSync('/app/certs/app.baselaragoproject.test.crt'),
      },
      cors: true,
      headers: {
        "Access-Control-Allow-Origin": "*",
      },
      hmr: {
        clientPort: env.VITE_HMR_PORT ? Number(env.VITE_HMR_PORT) : 443,
        host: env.VITE_VIRTUAL_HOST,
        protocol: 'wss'
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
      },
    },
  };
});