import { defineConfig } from '@vben/vite-config';

import ElementPlus from 'unplugin-element-plus/vite';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      plugins: [
        ElementPlus({
          format: 'esm',
        }),
      ],
      server: {
        host: '127.0.0.1',
        open: true,
        strictPort: true,
        proxy: {
          '/api': {
            changeOrigin: true,

            // 本地 Go API
            target: 'http://127.0.0.1:8080',
            ws: true,
          },
        },
      },
    },
  };
});
